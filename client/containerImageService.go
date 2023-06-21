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
	if err != nil {
		return err
	}

	return nil
}

// Delete container image
func (svc *ContainerImageService) Delete(ctx context.Context, image string) error {
	_, err := svc.client.Request(ctx, "container-image", map[string]any{
		"op":   "delete",
		"name": image,
	})
	if err != nil {
		return err
	}

	return nil
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
	if data == "" {
		return []ContainerImage{}, nil
	}
	lines := strings.Split(strings.TrimSpace(data), "\n")

	images := []ContainerImage{}
	j := len(lines)
	foundHeader := false
	for i := 0; i < j; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if !foundHeader {
			foundHeader = imageHeaderPattern.MatchString(line)
			continue
		}
		match := reSubMatchMap(imageLinePattern, line)
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

func reSubMatchMap(r *regexp.Regexp, str string) map[string]string {
	match := r.FindStringSubmatch(str)
	subMatchMap := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i != 0 {
			subMatchMap[name] = match[i]
		}
	}

	return subMatchMap
}
