package controller

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/crstrn13/internal-extensions/extensions/pipeline-policy/api/v1alpha1"
	extcontroller "github.com/kuadrant/kuadrant-operator/pkg/extension/controller"
	"github.com/kuadrant/kuadrant-operator/pkg/extension/types"
)

type PipelinePolicyReconciler struct {
	types.ExtensionBase
}

func NewPipelinePolicyReconciler() *PipelinePolicyReconciler {
	return &PipelinePolicyReconciler{}
}

func (r *PipelinePolicyReconciler) Reconcile(ctx context.Context, request reconcile.Request, kuadrantCtx types.KuadrantCtx) (reconcile.Result, error) {
	if err := r.Configure(ctx); err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to configure extension: %w", err)
	}
	r.Logger.Info("reconciling pipelinepolicy started")
	defer r.Logger.Info("reconciling pipelinepolicy completed")

	pol := &v1alpha1.PipelinePolicy{}
	if err := r.Client.Get(ctx, request.NamespacedName, pol); err != nil {
		if errors.IsNotFound(err) {
			r.Logger.Error(err, "pipelinepolicy not found")
			return reconcile.Result{}, nil
		}
		r.Logger.Error(err, "failed to retrieve pipelinepolicy")
		return reconcile.Result{}, err
	}

	if pol.GetDeletionTimestamp() != nil {
		r.Logger.Info("pipelinepolicy marked for deletion")
		return reconcile.Result{}, nil
	}

	policyStatus, specErr := r.reconcileSpec(ctx, pol, kuadrantCtx)
	statusResult, statusErr := r.reconcileStatus(ctx, pol, policyStatus)

	if specErr != nil {
		return reconcile.Result{}, specErr
	}
	if statusErr != nil {
		return reconcile.Result{}, statusErr
	}

	if statusResult.RequeueAfter > 0 {
		r.Logger.Info("Reconciling status not finished. Requeueing.")
		return statusResult, nil
	}

	return reconcile.Result{}, nil
}

func (r *PipelinePolicyReconciler) reconcileSpec(ctx context.Context, pol *v1alpha1.PipelinePolicy, kuadrantCtx types.KuadrantCtx) (*v1alpha1.PipelinePolicyStatus, error) {
	for _, am := range pol.Spec.ActionMethods {
		r.Logger.Info("registering action method", "name", am.Name, "url", am.URL)
		if err := kuadrantCtx.RegisterActionMethod(ctx, pol, types.ActionMethodConfig{
			Name:            am.Name,
			URL:             am.URL,
			Service:         am.Service,
			Method:          am.Method,
			MessageTemplate: am.MessageTemplate,
		}); err != nil {
			r.Logger.Error(err, "failed to register action method", "name", am.Name)
			return calculateErrorStatus(pol, err), err
		}
	}

	pipeline := kuadrantCtx.NewPipeline(pol)

	if err := validateFailActions(slices.Concat(pol.Spec.Request, pol.Spec.Response)); err != nil {
		return calculateErrorStatus(pol, err), err
	}

	requestActions, err := buildActions(pol.Spec.Request)
	if err != nil {
		return calculateErrorStatus(pol, err), err
	}
	if err := pipeline.OnHTTPRequest(requestActions...); err != nil {
		return calculateErrorStatus(pol, err), err
	}

	responseActions, err := buildActions(pol.Spec.Response)
	if err != nil {
		return calculateErrorStatus(pol, err), err
	}
	if err := pipeline.OnHTTPResponse(responseActions...); err != nil {
		return calculateErrorStatus(pol, err), err
	}

	if err := pipeline.Commit(ctx); err != nil {
		r.Logger.Error(err, "failed to commit pipeline")
		return calculateErrorStatus(pol, err), err
	}

	r.Logger.Info("pipeline committed successfully")
	return calculateEnforcedStatus(pol, nil), nil
}

func (r *PipelinePolicyReconciler) reconcileStatus(ctx context.Context, pol *v1alpha1.PipelinePolicy, newStatus *v1alpha1.PipelinePolicyStatus) (ctrl.Result, error) {
	equalStatus := pol.Status.Equals(newStatus, r.Logger)
	r.Logger.Info("Status", "status is different", !equalStatus)
	r.Logger.Info("Status", "generation is different", pol.Generation != pol.Status.ObservedGeneration)
	if equalStatus && pol.Generation == pol.Status.ObservedGeneration {
		r.Logger.Info("Status was not updated")
		return reconcile.Result{}, nil
	}

	r.Logger.V(1).Info("Updating Status", "sequence no:", fmt.Sprintf("sequence No: %v->%v", pol.Status.ObservedGeneration, newStatus.ObservedGeneration))

	pol.Status = *newStatus
	updateErr := r.Client.Status().Update(ctx, pol)
	if updateErr != nil {
		if errors.IsConflict(updateErr) {
			r.Logger.Info("Failed to update status: resource might just be outdated")
			return reconcile.Result{Requeue: true}, nil
		}
		return reconcile.Result{}, fmt.Errorf("failed to update status: %w", updateErr)
	}
	return ctrl.Result{}, nil
}

func calculateErrorStatus(pol *v1alpha1.PipelinePolicy, specErr error) *v1alpha1.PipelinePolicyStatus {
	newStatus := &v1alpha1.PipelinePolicyStatus{
		ObservedGeneration: pol.Generation,
		Conditions:         slices.Clone(pol.Status.Conditions),
	}
	meta.SetStatusCondition(&newStatus.Conditions, *extcontroller.AcceptedCondition(pol, specErr))
	meta.RemoveStatusCondition(&newStatus.Conditions, string(types.PolicyConditionEnforced))
	return newStatus
}

func calculateEnforcedStatus(pol *v1alpha1.PipelinePolicy, enforcedErr error) *v1alpha1.PipelinePolicyStatus {
	newStatus := &v1alpha1.PipelinePolicyStatus{
		ObservedGeneration: pol.Generation,
		Conditions:         slices.Clone(pol.Status.Conditions),
	}
	meta.SetStatusCondition(&newStatus.Conditions, *extcontroller.AcceptedCondition(pol, nil))
	meta.SetStatusCondition(&newStatus.Conditions, *extcontroller.EnforcedCondition(pol, enforcedErr, true))
	return newStatus
}

func buildActions(specs []v1alpha1.ActionSpec) ([]types.Action, error) {
	actions := make([]types.Action, 0, len(specs))
	for _, spec := range specs {
		switch spec.Type {
		case v1alpha1.ActionTypeGRPCMethod:
			actions = append(actions, types.GRPCMethodAction{
				Predicate: spec.Predicate,
				Method:    spec.Method,
				Var:       spec.Var,
			})
		case v1alpha1.ActionTypeDeny:
			actions = append(actions, types.DenyAction{
				Predicate:   spec.Predicate,
				WithStatus:  spec.WithStatus,
				WithHeaders: spec.WithHeaders,
				WithBody:    spec.WithBody,
			})
		case v1alpha1.ActionTypeFail:
			actions = append(actions, types.FailAction{
				Predicate:  spec.Predicate,
				LogMessage: spec.LogMessage,
			})
		case v1alpha1.ActionTypeAddHeaders:
			actions = append(actions, types.AddHeadersAction{
				Predicate:    spec.Predicate,
				HeadersToAdd: spec.HeadersToAdd,
			})
		default:
			return nil, fmt.Errorf("unknown action type: %s", spec.Type)
		}
	}
	return actions, nil
}

// validateFailActions checks that there are no top-level fail actions, because they will not work anyway.
func validateFailActions(specs []v1alpha1.ActionSpec) error {
	grpcVars := make([]string, 0)
	for _, spec := range specs {
		if spec.Var != "" {
			grpcVars = append(grpcVars, spec.Var)
		}
	}

	for _, spec := range specs {
		if spec.Type == v1alpha1.ActionTypeFail {
			flag := false
			for _, grpcVar := range grpcVars {
				if strings.Contains(spec.Predicate, grpcVar) || strings.Contains(spec.LogMessage, grpcVar) {
					flag = true
					break
				}
			}
			if !flag {
				return fmt.Errorf("top-level %s action types are not allowed", spec.Type)
			}
		}
	}

	return nil
}
