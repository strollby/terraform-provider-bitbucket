package bitbucket

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DrFaust92/bitbucket-go-client"
)

func handleClientError(httpResponse *http.Response, err error) error {
	if httpResponse == nil || httpResponse.StatusCode < 400 {
		return nil
	}

	clientHttpError, ok := err.(bitbucket.GenericSwaggerError)
	if ok {
		var errorBody string
		var bitbucketHttpError bitbucket.ModelError
		if err := json.Unmarshal(clientHttpError.Body(), &bitbucketHttpError); err != nil {
			errorBody = string(clientHttpError.Body()[:])
		} else {
			errorBody = bitbucketHttpError.Error_.Message
		}

		return fmt.Errorf("%s: %s", httpResponse.Status, errorBody)
	}

	if err != nil {
		return err
	}

	return nil
}
