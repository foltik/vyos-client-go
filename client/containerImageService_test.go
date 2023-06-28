package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit_ParseImages(t *testing.T) {
	// correct empty response
	images, err := parseImages("")
	assert.NoError(t, err, "expected no error parsing images")
	assert.Len(t, images, 0, "expected exactly 0 container images")

	header := "REPOSITORY                   TAG         IMAGE ID      CREATED       SIZE"
	name0, tag0, id0 := "docker.io/library/alpine0", "3.18.0.0", "5e2b554c1c450"
	name1, tag1, id1 := "docker.io/library/alpine1", "3.18.0.1", "5e2b554c1c451"
	image0 := name0 + "  " + tag0 + "  " + id0 + "  40 weeks ago  7.620 MB"
	image1 := name1 + "  " + tag1 + "  " + id1 + "  41 weeks ago  7.621 MB"

	// correct response
	images, err = parseImages(header + "\n" + image0 + "\n" + image1)
	assert.NoError(t, err, "expected no error parsing images")
	assert.Len(t, images, 2, "expected exactly 2 container images")
	assert.Equal(t, name0, images[0].Name, "image name must be equal")
	assert.Equal(t, tag0, images[0].Tag, "image tag must be equal")
	assert.Equal(t, id0, images[0].ImageID, "image id must be equal")
	assert.Equal(t, name1, images[1].Name, "image name must be equal")
	assert.Equal(t, tag1, images[1].Tag, "image tag must be equal")
	assert.Equal(t, id1, images[1].ImageID, "image id must be equal")

	// should ignore empty lines and lines preceding header
	images, err = parseImages("bogus\n \n\n" + image0 + "\n" + header + "\n" + image0 + "\n\n\n" + image1)
	assert.NoError(t, err, "expected no error parsing images")
	assert.Len(t, images, 2, "expected exactly 2 container images")

	// should error when no header present
	images, err = parseImages(image0 + "\n" + image1)
	assert.Error(t, err, "expected error parsing images")

	// should error on malformed image entry
	images, err = parseImages(header + "\n" + name0 + "  " + tag0)
	assert.Error(t, err, "expected error parsing images")
	images, err = parseImages(header + "\n" + "$")
	assert.Error(t, err, "expected error parsing images")
}

func TestIntegration_Containers_Show(t *testing.T) {
	client, ctx := make_client(t)

	// add image
	err := client.ContainerImages.Add(ctx, "alpine:3.17.3")
	assert.NoError(t, err, "expected no error adding container image")

	// show newly added image
	images, err := client.ContainerImages.Show(ctx)
	assert.NoError(t, err, "expected no error showing container images")

	// should find newly added image
	found := false
	for _, image := range images {
		if image.Name == "docker.io/library/alpine" && image.Tag == "3.17.3" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected to find container image for alpine:3.17.3")
}

func TestIntegration_Containers_Delete(t *testing.T) {
	client, ctx := make_client(t)

	// add image
	err := client.ContainerImages.Add(ctx, "alpine:3.17.3")
	assert.NoError(t, err, "expected no error adding container image")

	// delete image
	err = client.ContainerImages.Delete(ctx, "alpine:3.17.3")
	assert.NoError(t, err, "expected no error deleting container image")

	// show images
	images, err := client.ContainerImages.Show(ctx)
	assert.NoError(t, err, "expected no error showing container images")

	// should not find previously deleted image
	found := false
	for _, image := range images {
		if image.Name == "docker.io/library/alpine" && image.Tag == "3.17.3" {
			found = true
			break
		}
	}
	assert.False(t, found, "expected to NOT find container image for alpine:3.17.3")
}
