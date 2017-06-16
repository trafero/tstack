package tstackutil

import (
	"errors"
	"net/url"
)

// IsSecureUrl determines if the given URL has a secure scheme type
func IsSecureUrl(urlstring string) (bool, error) {
	u, err := url.Parse(urlstring)
	if err != nil {
		return false, err
	}

	switch u.Scheme {
	case "tcp":
		return false, nil
	case "ssl":
		return true, nil
	case "http":
		return false, nil
	case "https":
		return true, nil
	default:
		return false, errors.New("Unknown broker URL type.")
	}
}
