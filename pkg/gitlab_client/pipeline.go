package gitlab_client

import (

	"github.com/rs/zerolog/log"
	"github.com/xanzy/go-gitlab"
)

func (c *Client) GetPipelinesForCommit(projectName string, commitSHA string) ([]*gitlab.PipelineInfo, error) {
	pipelines, _, err := c.Pipelines.ListProjectPipelines(projectName, &gitlab.ListProjectPipelinesOptions{
		SHA: gitlab.String(commitSHA),
	})
	if err != nil {
		log.Error().Err(err).Msg("gitlab client: could not get pipelines for commit")
		return pipelines, err
	}

	return pipelines, nil

}

func (c *Client) GetLastPipelinesForCommit(projectName string, commitSHA string) *gitlab.PipelineInfo {
	pipelines, err := c.GetPipelinesForCommit(projectName, commitSHA)
	if err != nil {
		return nil
	}

	if len(pipelines) > 0 {
		// check for merge_requests_event
		for _, p := range pipelines {
			if p.Source == "merge_request_event" {
				return p
			}
		}

		// check for external_pull_request_events next
		for _, p := range pipelines {
			if p.Source == "pipeline" {
				return p
			}
		}

		// check for external_pull_request_events next
		for _, p := range pipelines {
			if p.Source == "external_pull_request_event" {
				return p
			}
		}

		for _, p := range pipelines {
			if p.Source == "external" {
				return p
			}
		}
	}

	return nil
}
