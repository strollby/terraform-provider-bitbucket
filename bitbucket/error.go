package bitbucket

import (
	"encoding/json"
	"fmt"

	"github.com/DrFaust92/bitbucket-go-client"
)

func handleClientError(err error) error {
	httpErr, ok := err.(bitbucket.GenericSwaggerError)
	if ok {
		var httpError bitbucket.ModelError
		if err := json.Unmarshal(httpErr.Body(), &httpError); err != nil {
			return fmt.Errorf(string(httpErr.Body()))
		}

		return fmt.Errorf("%s: %s", httpErr.Error(), httpError.Error_.Message)
	}

	if err != nil {
		return err
	}

	return nil
}
