package integrate

import (
	"testing"

	"github.com/itchio/butler/butlerd"
	"github.com/itchio/butler/butlerd/messages"
	"github.com/itchio/mitch"
)

func Test_ChangeCase(t *testing.T) {
	bi := newInstance(t)
	rc, _, cancel := bi.Unwrap()
	defer cancel()

	bi.Authenticate()

	store := bi.Server.Store()
	_developer := store.MakeUser("John Doe")
	_game := _developer.MakeGame("Airplane Simulator")
	_game.Type = "html"
	_game.Publish()
	_upload := _game.MakeUpload("All platforms")

	_upload.SetAllPlatforms()
	_upload.PushBuild(func(ac *mitch.ArchiveContext) {
		ac.Entry("index.html").String("<p>Hi!</p>")
		ac.Entry("data/data1").Random(0x1, 1024)
		ac.Entry("data/data2").Random(0x2, 1024)
		ac.Entry("data/data3").Random(0x3, 1024)
	})

	game := bi.FetchGame(_game.ID)

	queueRes, err := messages.InstallQueue.TestCall(rc, butlerd.InstallQueueParams{
		Game:              game,
		InstallLocationID: "tmp",
	})
	must(err)

	_, err = messages.InstallPerform.TestCall(rc, butlerd.InstallPerformParams{
		ID:            queueRes.ID,
		StagingFolder: queueRes.StagingFolder,
	})
	must(err)

	bi.Logf("Pushing second build...")

	_build2 := _upload.PushBuild(func(ac *mitch.ArchiveContext) {
		ac.Entry("index.html").String("<p>Hi!</p>")
		ac.Entry("Data/data1").Random(0x1, 1024)
		ac.Entry("Data/data2").Random(0x2, 1024)
		ac.Entry("Data/data3").Random(0x3, 1024)
	})

	bi.Logf("Now upgrading to second build...")

	caveId := queueRes.CaveID
	upload := bi.FetchUpload(_upload.ID)
	build := bi.FetchBuild(_build2.ID)

	queueRes, err = messages.InstallQueue.TestCall(rc, butlerd.InstallQueueParams{
		Game: game,
		// make sure to install to same cave so that it ends up
		// being an upgrade and not a duplicate install
		CaveID:            caveId,
		InstallLocationID: "tmp",

		// force upgrade otherwise it's going to default
		// to reinstall
		Upload: upload,
		Build:  build,
	})
	must(err)

	_, err = messages.InstallPerform.TestCall(rc, butlerd.InstallPerformParams{
		ID:            queueRes.ID,
		StagingFolder: queueRes.StagingFolder,
	})
	must(err)

	bi.Logf("Now re-install (heal)")

	queueRes, err = messages.InstallQueue.TestCall(rc, butlerd.InstallQueueParams{
		Game:              game,
		CaveID:            caveId,
		InstallLocationID: "tmp",
	})
	must(err)

	_, err = messages.InstallPerform.TestCall(rc, butlerd.InstallPerformParams{
		ID:            queueRes.ID,
		StagingFolder: queueRes.StagingFolder,
	})
	must(err)
}