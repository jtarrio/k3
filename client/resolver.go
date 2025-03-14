package client

import "context"

// HandleResolver returns a function that takes a handle and returns its DID, if it exists.
//
// You can use this function with importers to link usernames to their profiles.
func HandleResolver(client Client) func(handle string) *string {
	return func(handle string) *string {
		result, err := client.FindUserByHandle(context.Background(), handle)
		if err != nil {
			return nil
		}
		return &result.Did
	}
}
