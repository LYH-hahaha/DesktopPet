//go:build windows

package platform

func NewPlatform() (Platform, error) {
	return &windowsPlatform{}, nil
}
