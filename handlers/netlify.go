package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/djatwood/formailer"
	"github.com/djatwood/formailer/logger"
)

func netlifyResponse(code int, err error, headers ...[2]string) *events.APIGatewayProxyResponse {
	r := response{}
	response := &events.APIGatewayProxyResponse{
		StatusCode: code,
		Headers:    make(map[string]string),
	}

	for _, h := range headers {
		response.Headers[h[0]] = h[1]
	}

	if err != nil {
		r.Ok = false
		r.Error = err.Error()
		if _, ok := response.Headers["location"]; ok {
			response.Headers["location"] += "?error=" + err.Error()
		}
		logger.Error(err)
	}

	body, err := json.Marshal(r)
	if err != nil {
		logger.Errorf("failed to marshal response: %w", err)
	}

	response.Body = string(body)
	return response
}

// Netlify takes in a aws lambda request and sends an email
func Netlify(c formailer.Config) func(events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	return func(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
		if request.HTTPMethod != "POST" {
			return netlifyResponse(http.StatusMethodNotAllowed, nil), nil
		}

		submission, err := formailer.Parse(request.Headers["content-type"], request.Body)
		if err != nil {
			return netlifyResponse(http.StatusBadRequest, err), nil
		}

		if submission.Form.ReCAPTCHA {
			v, exists := submission.Values["g-recaptcha-response"].(string)
			if !exists || len(v) < 1 {
				return netlifyResponse(http.StatusBadRequest, err), nil
			}

			ok, err := VerifyRecaptcha(v)
			if err != nil {
				return netlifyResponse(http.StatusInternalServerError, err), nil
			}
			if !ok {
				return netlifyResponse(http.StatusBadRequest, nil), nil
			}

			delete(submission.Values, "g-recaptcha-response")
		}

		for _, email := range submission.Emails {
			email.To = ReplaceDynamic(email.To, submission)
			email.Subject = ReplaceDynamic(email.Subject, submission)
		}

		err = submission.Send()
		if err != nil {
			err = fmt.Errorf("failed to send email: %w", err)
			return netlifyResponse(http.StatusInternalServerError, err), nil
		}

		statusCode := http.StatusOK
		headers := [][2]string{}
		if len(submission.Form.Redirect) > 0 {
			statusCode = http.StatusSeeOther
			headers = append(headers, [2]string{"location", submission.Form.Redirect})
		}

		logger.Infof("sent %d emails from %s form", len(submission.Form.Emails), submission.Values["_form_name"])
		return netlifyResponse(statusCode, nil, headers...), nil
	}
}
