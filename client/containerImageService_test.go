package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit_ParseImages(t *testing.T) {
	images, err := parseImages("sudo: unable to resolve host 62b1b7ccea6c: System error\nREPOSITORY                   TAG         IMAGE ID      CREATED       SIZE\ndocker.io/library/alpine     3.18.0      5e2b554c1c45  4 weeks ago   7.62 MB\n")
	assert.NoError(t, err, "expected no error parsing images")
	assert.Len(t, images, 1, "expected exactly 1 container image")
	assert.Equal(t, "docker.io/library/alpine", images[0].Name, "image name must be equal")
	assert.Equal(t, "3.18.0", images[0].Tag, "image tag must be equal")
	assert.Equal(t, "5e2b554c1c45", images[0].ImageID, "image id must be equal")
}

func TestIntegration_Containers_Show(t *testing.T) {
	client, ctx := make_client(t)
	err := client.ContainerImages.Add(ctx, "alpine:3.17.3")
	assert.NoError(t, err, "expected no error adding container image")
	images, err := client.ContainerImages.Show(ctx)
	assert.NoError(t, err, "expected no error showing container images")
	var actual *ContainerImage = nil
	for _, image := range images {
		if image.Name == "docker.io/library/alpine" && image.Tag == "3.17.3" {
			actual = &image
		}
	}
	assert.NotNil(t, actual, "expected to find container image for alpine:3.17.3")
}

func TestIntegration_Containers_Delete(t *testing.T) {
	client, ctx := make_client(t)
	err := client.ContainerImages.Add(ctx, "alpine:3.17.3")
	assert.NoError(t, err, "expected no error adding container image")
	err = client.ContainerImages.Delete(ctx, "alpine:3.17.3")
	assert.NoError(t, err, "expected no error deleting container image")

	images, err := client.ContainerImages.Show(ctx)
	assert.NoError(t, err, "expected no error showing container images")
	var actual *ContainerImage = nil
	for _, image := range images {
		if image.Name == "docker.io/library/alpine" && image.Tag == "3.17.3" {
			actual = &image
		}
	}
	assert.Nil(t, actual, "expected to NOT find container image for alpine:3.17.3")
}
