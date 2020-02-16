/*
 * Copyright (C) 2016-2018. ActionTech.
 * Based on: github.com/actiontech/dtle, github.com/github/gh-ost .
 * License: MPL version 2: https://www.mozilla.org/en-US/MPL/2.0 .
 */

package server

import (
	"time"

	"github.com/actiontech/dtle/olddtle/internalinternal/models"
)

// Plan endpoint is used for plan interactions
type Plan struct {
	srv *Server
}

// Submit is used to submit a plan to the leader
func (p *Plan) Submit(args *models.PlanRequest, reply *models.PlanResponse) error {
	if done, err := p.srv.forward("Plan.Submit", args, args, reply); done {
		return err
	}
	defer metrics.MeasureSince([]string{"server", "plan", "submit"}, time.Now())

	// Pause the Nack timer for the eval as it is making progress as long as it
	// is in the plan queue. We resume immediately after we get a result to
	// handle the case that the receiving worker dies.
	plan := args.Plan
	id := plan.EvalID
	token := plan.EvalToken
	if err := p.srv.evalBroker.PauseNackTimeout(id, token); err != nil {
		return err
	}
	defer p.srv.evalBroker.ResumeNackTimeout(id, token)

	// Submit the plan to the queue
	future, err := p.srv.planQueue.Enqueue(plan)
	if err != nil {
		return err
	}

	// Wait for the results
	result, err := future.Wait()
	if err != nil {
		return err
	}

	// Package the result
	reply.Result = result
	reply.Index = result.AllocIndex
	return nil
}
