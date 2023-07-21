package config

import (
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/rs/zerolog/log"
	"github.com/zapier/kubechecks/pkg/app_directory"
)

type VcsToArgoMap struct {
	vcsAppStubsByRepo map[repoURL]*app_directory.AppDirectory
}

func NewVcsToArgoMap() VcsToArgoMap {
	return VcsToArgoMap{
		vcsAppStubsByRepo: make(map[repoURL]*app_directory.AppDirectory),
	}
}

func (v2a *VcsToArgoMap) GetAppsInRepo(repoCloneUrl string) *app_directory.AppDirectory {
	repoUrl, err := normalizeRepoUrl(repoCloneUrl)
	if err != nil {
		log.Warn().Err(err).Msgf("failed to parse %s", repoCloneUrl)
	}

	return v2a.vcsAppStubsByRepo[repoUrl]
}

func (v2a *VcsToArgoMap) AddApp(app *v1alpha1.Application) {
	if app.Spec.Source == nil {
		return
	}

	rawRepoUrl := app.Spec.Source.RepoURL
	cleanRepoUrl, err := normalizeRepoUrl(rawRepoUrl)
	if err != nil {
		log.Warn().Err(err).Msgf("failed to parse %s", rawRepoUrl)
		return
	}

	appDirectory := v2a.vcsAppStubsByRepo[cleanRepoUrl]
	if appDirectory == nil {
		appDirectory = app_directory.NewAppDirectory()
	}
	appDirectory.AddApp(app)
	v2a.vcsAppStubsByRepo[cleanRepoUrl] = appDirectory
}
