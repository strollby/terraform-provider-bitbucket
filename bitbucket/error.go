package bitbucket

import (
	"encoding/json"

	"github.com/DrFaust92/bitbucket-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func handleClientError(err error) diag.Diagnostics {
	httpErr, ok := err.(bitbucket.GenericSwaggerError)
	if ok {
		var httpError bitbucket.ModelError
		if err := json.Unmarshal(httpErr.Body(), &httpError); err != nil {
			return diag.Errorf(string(httpErr.Body()))
		}

		return diag.Errorf("%s: %s", httpErr.Error(), httpError.Error_.Message)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
