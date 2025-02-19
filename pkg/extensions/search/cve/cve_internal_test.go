//go:build search

package cveinfo

import (
	"errors"
	"testing"
	"time"

	"github.com/opencontainers/go-digest"
	ispec "github.com/opencontainers/image-spec/specs-go/v1"
	. "github.com/smartystreets/goconvey/convey"

	cvemodel "zotregistry.io/zot/pkg/extensions/search/cve/model"
	"zotregistry.io/zot/pkg/meta/types"
	"zotregistry.io/zot/pkg/test/mocks"
)

var ErrTestError = errors.New("test error")

func TestUtils(t *testing.T) {
	Convey("Utils", t, func() {
		Convey("getConfigAndDigest", func() {
			_, _, err := getConfigAndDigest(mocks.MetaDBMock{}, "bad-digest")
			So(err, ShouldNotBeNil)

			_, _, err = getConfigAndDigest(mocks.MetaDBMock{
				GetImageMetaFn: func(digest digest.Digest) (types.ImageMeta, error) {
					return types.ImageMeta{}, ErrTestError
				},
			}, ispec.DescriptorEmptyJSON.Digest.String())
			So(err, ShouldNotBeNil)

			// bad media type of config
			_, _, err = getConfigAndDigest(mocks.MetaDBMock{
				GetImageMetaFn: func(digest digest.Digest) (types.ImageMeta, error) {
					return types.ImageMeta{Manifests: []types.ManifestMeta{
						{Manifest: ispec.Manifest{Config: ispec.Descriptor{MediaType: "bad-type"}}},
					}}, nil
				},
			}, ispec.DescriptorEmptyJSON.Digest.String())
			So(err, ShouldNotBeNil)
		})
		Convey("getIndexContent", func() {
			_, err := getIndexContent(mocks.MetaDBMock{}, "bad-digest")
			So(err, ShouldNotBeNil)

			_, err = getIndexContent(mocks.MetaDBMock{
				GetImageMetaFn: func(digest digest.Digest) (types.ImageMeta, error) {
					return types.ImageMeta{}, ErrTestError
				},
			}, ispec.DescriptorEmptyJSON.Digest.String())
			So(err, ShouldNotBeNil)

			// nil index
			_, err = getIndexContent(mocks.MetaDBMock{
				GetImageMetaFn: func(digest digest.Digest) (types.ImageMeta, error) {
					return types.ImageMeta{}, nil
				},
			}, ispec.DescriptorEmptyJSON.Digest.String())
			So(err, ShouldNotBeNil)
		})

		Convey("mostRecentUpdate", func() {
			// empty
			timestamp := mostRecentUpdate([]cvemodel.DescriptorInfo{})
			So(timestamp, ShouldResemble, time.Time{})

			timestamp = mostRecentUpdate([]cvemodel.DescriptorInfo{
				{
					Timestamp: time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC),
				},
				{
					Timestamp: time.Date(2005, 1, 1, 1, 1, 1, 1, time.UTC),
				},
			})
			So(timestamp, ShouldResemble, time.Date(2005, 1, 1, 1, 1, 1, 1, time.UTC))
		})

		Convey("GetFixedTags", func() {
			tags := GetFixedTags(
				[]cvemodel.TagInfo{
					{},
				},
				[]cvemodel.TagInfo{
					{
						Descriptor: cvemodel.Descriptor{
							MediaType: ispec.MediaTypeImageManifest,
						},
						Timestamp: time.Date(2010, 1, 1, 1, 1, 1, 1, time.UTC),
					},
					{
						Descriptor: cvemodel.Descriptor{
							MediaType: ispec.MediaTypeImageIndex,
						},
						Manifests: []cvemodel.DescriptorInfo{
							{
								Timestamp: time.Date(2002, 1, 1, 1, 1, 1, 1, time.UTC),
							},
							{
								Timestamp: time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC),
							},
						},
					},
					{
						Descriptor: cvemodel.Descriptor{
							MediaType: "bad Type",
						},
					},
				})
			So(tags, ShouldBeEmpty)
		})
	})
}
