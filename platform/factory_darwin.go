//go:build darwin

package platform

func NewPlatform() (Platform, error) {
	return &darwinPlatform{}, nil
}
