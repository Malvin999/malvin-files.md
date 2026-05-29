package feishu

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMessageResourceExtensionUsesContentTypeForImagesWithoutFilename(t *testing.T) {
	require.Equal(t, ".jpg", messageResourceExtension("image", ".png", "", "image/jpeg"))
	require.Equal(t, ".webp", messageResourceExtension("image", ".png", "", "image/webp; charset=utf-8"))
}

func TestMessageResourceExtensionPrefersFilename(t *testing.T) {
	require.Equal(t, ".gif", messageResourceExtension("image", ".png", "photo.gif", "image/jpeg"))
}

func TestMessageResourceExtensionFallsBackForUnknownImages(t *testing.T) {
	require.Equal(t, ".png", messageResourceExtension("image", "", "", "application/octet-stream"))
}
