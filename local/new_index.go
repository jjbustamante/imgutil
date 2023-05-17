package local

import (
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/types"

	"github.com/buildpacks/imgutil"
)

func NewIndex(repoName string, path string, ops ...ImageIndexOption) (*ImageIndex, error) {
	if _, err := name.ParseReference(repoName, name.WeakValidation); err != nil {
		return nil, err
	}

	indexOpts := &indexOptions{}
	for _, op := range ops {
		if err := op(indexOpts); err != nil {
			return nil, err
		}
	}

	if len(indexOpts.manifest.Manifests) != 0 {
		index, err := emptyIndex(indexOpts.manifest.MediaType)
		if err != nil {
			return nil, err
		}

		for _, manifest_i := range indexOpts.manifest.Manifests {
			img, _ := emptyImage(imgutil.Platform{
				Architecture: manifest_i.Platform.Architecture,
				OS:           manifest_i.Platform.OS,
				OSVersion:    manifest_i.Platform.OSVersion,
			})
			index = mutate.AppendManifests(index, mutate.IndexAddendum{Add: img, Descriptor: manifest_i})
		}

		idx := &ImageIndex{
			repoName: repoName,
			path:     path,
			index:    index,
		}

		return idx, nil

	}

	mediaType := defaultMediaType()
	if indexOpts.mediaTypes.IndexManifestType() != "" {
		mediaType = indexOpts.mediaTypes
	}

	index, err := emptyIndex(mediaType.IndexManifestType())
	if err != nil {
		return nil, err
	}

	ridx := &ImageIndex{
		repoName: repoName,
		path:     path,
		index:    index,
	}

	return ridx, nil

}

func emptyIndex(mediaType types.MediaType) (v1.ImageIndex, error) {
	return mutate.IndexMediaType(empty.Index, mediaType), nil
}

func emptyImage(platform imgutil.Platform) (v1.Image, error) {
	cfg := &v1.ConfigFile{
		Architecture: platform.Architecture,
		OS:           platform.OS,
		OSVersion:    platform.OSVersion,
		RootFS: v1.RootFS{
			Type:    "layers",
			DiffIDs: []v1.Hash{},
		},
	}

	return mutate.ConfigFile(empty.Image, cfg)
}

func defaultMediaType() imgutil.MediaTypes {
	return imgutil.DockerTypes
}