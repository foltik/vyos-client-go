package client

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type ContainerImageService struct{ client *Client }
type ContainerImage struct {
	Name    string
	Tag     string
	ImageID string
}

// Add container image
func (svc *ContainerImageService) Add(ctx context.Context, image string) error {
	_, err := svc.client.Request(ctx, "container-image", map[string]any{
		"op":   "add",
		"name": image,
	})
	return err
}

// Delete container image
func (svc *ContainerImageService) Delete(ctx context.Context, image string) error {
	_, err := svc.client.Request(ctx, "container-image", map[string]any{
		"op":   "delete",
		"name": image,
	})
	return err
}

// Return the list of container images
func (svc *ContainerImageService) Show(ctx context.Context) ([]ContainerImage, error) {
	resp, err := svc.client.Request(ctx, "container-image", map[string]any{
		"op": "show",
	})
	if err != nil {
		return nil, err
	}

	data, ok := resp.(string)
	if !ok {
		return nil, errors.New("received unexpected repsonse format from server")
	}

	return parseImages(data)
}

var imageHeaderPattern = regexp.MustCompile(`^REPOSITORY\s{2,}TAG\s{2,}IMAGE ID\s{2,}.*$`)
var imageLinePattern = regexp.MustCompile(`^(?P<name>[^\s]+)\s{2,}(?P<tag>[^\s]+)\s{2,}(?P<imageId>[^\s]+)`)

func parseImages(data string) ([]ContainerImage, error) {
	data = strings.TrimSpace(data)
	if data == "" {
		return []ContainerImage{}, nil
	}

	images := []ContainerImage{}

	foundHeader := false
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if !foundHeader {
			foundHeader = imageHeaderPattern.MatchString(line)
			continue
		}

		match, ok := matchStringNamed(imageLinePattern, line)
		if !ok {
			return nil, fmt.Errorf("invalid image in response from vyos api:\n%s", line)
		}

		images = append(images, ContainerImage{
			Name:    match["name"],
			Tag:     match["tag"],
			ImageID: match["imageId"],
		})
	}

	if !foundHeader {
		return nil, fmt.Errorf("could not find expected container image header in response from vyos api:\n%s", data)
	}
	return images, nil
}

func matchStringNamed(r *regexp.Regexp, str string) (map[string]string, bool) {
	match := r.FindStringSubmatch(str)
	if match == nil || len(match) != len(r.SubexpNames()) {
		return nil, false
	}

	matches := map[string]string{}
	for i, name := range r.SubexpNames()[1:] {
		matches[name] = match[i+1]
	}
	return matches, true
}
