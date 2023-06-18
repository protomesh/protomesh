package aws

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/protomesh/go-app"
	typesv1 "github.com/protomesh/protomesh/proto/api/types/v1"
)

type lambdaHttpHandler struct {
	log app.Logger

	param *typesv1.AwsHandler_LambdaFunction

	lambdaCli *lambda.Client

	ctx context.Context

	request  *http.Request
	response http.ResponseWriter
}

func (l *lambdaHttpHandler) Call() error {

	reqBody, err := ioutil.ReadAll(l.request.Body)
	if err != nil {
		return err
	}

	req := &events.APIGatewayProxyRequest{
		HTTPMethod:                      l.request.Method,
		Path:                            l.request.URL.Path,
		MultiValueHeaders:               l.request.Header,
		MultiValueQueryStringParameters: l.request.URL.Query(),
		Body:                            string(reqBody[:]),
		IsBase64Encoded:                 false,
	}

	in, err := json.Marshal(req)
	if err != nil {
		return err
	}

	out, err := l.lambdaCli.Invoke(l.ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(l.param.FunctionName),
		InvocationType: types.InvocationTypeRequestResponse,
		Qualifier:      aws.String(l.param.Qualifier),
		Payload:        in,
	})
	if out.FunctionError != nil {
		return fmt.Errorf("Lambda function returned error: %s", aws.ToString(out.FunctionError))
	}
	if err != nil {
		return err
	}

	res := &events.APIGatewayProxyResponse{}
	if err := json.Unmarshal(out.Payload, res); err != nil {
		return err
	}

	if res.MultiValueHeaders != nil {
		for key, vals := range res.MultiValueHeaders {
			l.response.Header()[key] = vals
		}
	}

	if res.Headers != nil {
		for key, val := range res.Headers {
			l.response.Header().Add(key, val)
		}
	}

	l.response.WriteHeader(res.StatusCode)

	if res.IsBase64Encoded {

		result, err := base64.RawStdEncoding.DecodeString(res.Body)
		if err != nil {
			return err
		}

		if _, err := l.response.Write(result); err != nil {
			return err
		}

		return io.EOF

	}

	if _, err := l.response.Write([]byte(res.Body)); err != nil {
		return err
	}

	return io.EOF

}
