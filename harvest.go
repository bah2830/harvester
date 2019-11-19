package main

import (
	"context"
	"errors"
	"time"

	jira "github.com/andygrunwald/go-jira"
	"github.com/becoded/go-harvest/harvest"
	"golang.org/x/oauth2"
)

type harvestTask struct {
	harvest.UserProjectAssignment
	timer *harvest.TimeEntry
}

func (h *harvester) getNewHarvestClient() error {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: h.settings.Harvest.Pass,
		},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	service := harvest.NewHarvestClient(tc)
	service.AccountId = h.settings.Harvest.User

	h.harvestClient = service
	return nil
}

func (h *harvester) getHarvestCompany() (*harvest.Company, error) {
	ctx, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()

	company, _, err := h.harvestClient.Company.Get(ctx)
	if err != nil {
		return nil, err
	}

	return company, nil
}

func (h *harvester) getHarvestProjects() ([]*harvestTask, error) {
	ctx, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()

	asignments, _, err := h.harvestClient.Project.GetMyProjectAssignments(ctx, nil)
	if err != nil {
		return nil, err
	}

	tasks := make([]*harvestTask, len(asignments.UserAssignments))
	for i, a := range asignments.UserAssignments {
		task := &harvestTask{
			UserProjectAssignment: *a,
		}

		tasks[i] = task
	}

	timers, err := h.getTimers()
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		for _, timer := range timers {
			if *timer.Project.Id == *task.Project.Id {
				task.timer = timer
				break
			}
		}
	}
	return tasks, nil
}

func (h *harvester) getTimers() ([]*harvest.TimeEntry, error) {
	ctx, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()

	isRunning := true
	times, _, err := h.harvestClient.Timesheet.List(ctx, &harvest.TimeEntryListOptions{
		IsRunning: &isRunning,
	})
	if err != nil {
		return nil, err
	}

	return times.TimeEntries, nil
}

func (h *harvester) getMatchingHarvestProject(jira jira.Issue) *harvestTask {
	for _, p := range h.activeHarvest {
		if *p.Project.Code == jira.Key {
			return p
		}
	}

	return nil
}

func (h *harvester) startTimer(task *harvestTask) error {
	if task.timer != nil {
		return nil
	}

	// Get the coding task
	var codingTask *harvest.ProjectTaskAssignment
	for _, t := range *task.TaskAssignments {
		if *t.Task.Name == "Coding" {
			codingTask = &t
			break
		}
	}
	if codingTask == nil {
		return errors.New("unable to find coding task")
	}

	ctx, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()

	t, _, err := h.harvestClient.Timesheet.CreateTimeEntryViaDuration(ctx, &harvest.TimeEntryCreateViaDuration{
		ProjectId: task.Project.Id,
		TaskId:    codingTask.Task.Id,
		SpentDate: &harvest.Date{Time: time.Now()},
	})
	if err != nil {
		return err
	}

	task.timer = t
	return nil
}

func (h *harvester) stopTimer(task *harvestTask) error {
	if task.timer == nil {
		return nil
	}

	ctx, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()

	_, _, err := h.harvestClient.Timesheet.StopTimeEntry(ctx, *task.timer.Id)

	task.timer = nil
	return err
}
