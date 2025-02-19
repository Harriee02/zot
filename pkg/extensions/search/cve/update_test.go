//go:build search
// +build search

package cveinfo_test

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	ispec "github.com/opencontainers/image-spec/specs-go/v1"
	. "github.com/smartystreets/goconvey/convey"

	"zotregistry.io/zot/pkg/api/config"
	cveinfo "zotregistry.io/zot/pkg/extensions/search/cve"
	"zotregistry.io/zot/pkg/log"
	mTypes "zotregistry.io/zot/pkg/meta/types"
	"zotregistry.io/zot/pkg/scheduler"
	"zotregistry.io/zot/pkg/storage"
	test "zotregistry.io/zot/pkg/test/common"
	"zotregistry.io/zot/pkg/test/mocks"
)

func TestCVEDBGenerator(t *testing.T) {
	Convey("Test CVE DB task scheduler reset", t, func() {
		logFile, err := os.CreateTemp(t.TempDir(), "zot-log*.txt")
		logPath := logFile.Name()
		So(err, ShouldBeNil)

		defer os.Remove(logFile.Name()) // clean up

		logger := log.NewLogger("debug", logPath)
		writers := io.MultiWriter(os.Stdout, logFile)
		logger.Logger = logger.Output(writers)

		cfg := config.New()
		cfg.Scheduler = &config.SchedulerConfig{NumWorkers: 3}
		sch := scheduler.NewScheduler(cfg, logger)

		metaDB := &mocks.MetaDBMock{
			GetRepoMetaFn: func(ctx context.Context, repo string) (mTypes.RepoMeta, error) {
				return mTypes.RepoMeta{
					Tags: map[string]mTypes.Descriptor{
						"tag": {MediaType: ispec.MediaTypeImageIndex},
					},
				}, nil
			},
		}
		storeController := storage.StoreController{
			DefaultStore: mocks.MockedImageStore{
				RootDirFn: func() string {
					return t.TempDir()
				},
			},
		}

		cveScanner := cveinfo.NewScanner(storeController, metaDB, "ghcr.io/project-zot/trivy-db", "", logger)
		generator := cveinfo.NewDBUpdateTaskGenerator(time.Minute, cveScanner, logger)

		sch.SubmitGenerator(generator, 12000*time.Millisecond, scheduler.HighPriority)

		ctx, cancel := context.WithCancel(context.Background())

		sch.RunScheduler(ctx)

		defer cancel()

		// Wait for trivy db to download
		found, err := test.ReadLogFileAndCountStringOccurence(logPath,
			"DB update completed, next update scheduled", 140*time.Second, 2)
		So(err, ShouldBeNil)
		So(found, ShouldBeTrue)
	})
}
