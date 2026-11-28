package handlers

import (
	"net/http"
	"strconv"

	"plan/api/internal/models"
	"plan/api/internal/repository"

	"github.com/gin-gonic/gin"
)

// GoalVerificationHandler hanterar verifieringsplaner, bevis och omdömen.
//
// Planen är valfri - ett mål utan plan fungerar precis som ett mål med en
// plan. När en plan finns blir den villkoret för att agenten får föreslå att
// målet är klart, och gränssnittet visar beviset innan människan bestämmer.
type GoalVerificationHandler struct {
	verification *repository.GoalVerificationRepository
	goals        *repository.GoalRepository
	dbConnected  bool
}

// NewGoalVerificationHandler ger en handler för verifieringar.
func NewGoalVerificationHandler(dbConnected bool) *GoalVerificationHandler {
	return &GoalVerificationHandler{
		verification: repository.NewGoalVerificationRepository(),
		goals:        repository.NewGoalRepository(),
		dbConnected:  dbConnected,
	}
}

// GetVerification godoc
// @Summary Get a goal's verification plan
// @Description Returns the goal's verification plan with its latest outcome and evidence. A goal without a plan returns an empty plan.
// @Tags goals
// @Produce json
// @Security BearerAuth
// @Param id path int true "Goal ID"
// @Success 200 {object} models.GoalVerification
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/goals/{id}/verification [get]
func (h *GoalVerificationHandler) GetVerification(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	_, goalID, ok := h.goalAccess(c)
	if !ok {
		return
	}

	plan, err := h.verification.GetByGoal(c.Request.Context(), goalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	evidence, err := h.verification.ListEvidence(c.Request.Context(), goalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if evidence == nil {
		evidence = []models.GoalEvidence{}
	}

	if plan == nil {
		// Ett mål utan plan svarar med en tom plan - inte ett fel. Så
		// kan klienten visa samma vy hela tiden.
		plan = &models.GoalVerification{GoalID: goalID}
	}

	result, err := h.verification.LatestResult(c.Request.Context(), goalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if result != nil {
		plan.Outcome = result.Outcome
		plan.DecidedBy = result.DecidedBy
		plan.DecidedAt = result.DecidedAt
	}

	svarbody := gin.H{
		"verification": plan,
		"evidence":     evidence,
	}
	if result != nil {
		// Omdömet skickas med i sin helhet - annars kan klienten inte
		// godkänna det (godkännandet gäller ett resultat, inte ett
		// utfall).
		svarbody["result"] = result
	}
	c.JSON(http.StatusOK, svarbody)
}

// UpdateVerification godoc
// @Summary Set a goal's verification plan
// @Description Sets the verification plan for a goal. The plan is optional - changing it invalidates earlier outcomes.
// @Tags goals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Goal ID"
// @Param request body models.UpdateGoalVerificationRequest true "Verification plan"
// @Success 200 {object} models.GoalVerification
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/goals/{id}/verification [put]
func (h *GoalVerificationHandler) UpdateVerification(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	_, goalID, ok := h.goalAccess(c)
	if !ok {
		return
	}

	var req models.UpdateGoalVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "type and criterion are required"})
		return
	}

	if !validVerificationType(req.Type) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "unknown verification type"})
		return
	}

	strength := req.Strength
	if strength == "" {
		strength = models.VerificationStrengthWeak
	}
	if strength != models.VerificationStrengthStrong && strength != models.VerificationStrengthWeak {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "unknown verification strength"})
		return
	}

	plan, err := h.verification.Upsert(c.Request.Context(), goalID, req.Type, req.Criterion, strength)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, plan)
}

// DeleteVerification tar bort verifieringsplanen från ett mål.
//
// Planen är valfri - att ta bort den lämnar målet som det var, bara
// utan verifiering.
// @Summary Remove the verification plan from a goal
// @Description Removes the verification plan. The goal keeps working - the plan is optional.
// @Tags goals
// @Produce json
// @Security BearerAuth
// @Param id path int true "Goal ID"
// @Success 204
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/goals/{id}/verification [delete]
func (h *GoalVerificationHandler) DeleteVerification(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	_, goalID, ok := h.goalAccess(c)
	if !ok {
		return
	}

	if err := h.verification.Delete(c.Request.Context(), goalID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// AddEvidence godoc
// @Summary Add evidence to a goal
// @Description Stores evidence for a goal - the proof that a goal was actually reached.
// @Tags goals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Goal ID"
// @Param request body models.CreateGoalEvidenceRequest true "Evidence"
// @Success 201 {object} models.GoalEvidence
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/goals/{id}/evidence [post]
func (h *GoalVerificationHandler) AddEvidence(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	_, goalID, ok := h.goalAccess(c)
	if !ok {
		return
	}

	var req models.CreateGoalEvidenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "type and value are required"})
		return
	}

	var verificationID *int64
	plan, err := h.verification.GetByGoal(c.Request.Context(), goalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if plan != nil {
		verificationID = &plan.ID
	}

	source := req.Source
	if source == "" {
		source = "user"
	}

	evidence, err := h.verification.AddEvidence(c.Request.Context(), goalID, verificationID, req.Type, req.Value, source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, evidence)
}

// RunVerification godoc
// @Summary Run a goal's verification
// @Description Runs the goal's verification and stores the outcome. The outcome is not approved - a person has to approve it before it counts as verified.
// @Tags goals
// @Produce json
// @Security BearerAuth
// @Param id path int true "Goal ID"
// @Success 200 {object} models.GoalVerificationResult
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/goals/{id}/verification/run [post]
func (h *GoalVerificationHandler) RunVerification(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	_, goalID, ok := h.goalAccess(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()

	plan, err := h.verification.GetByGoal(ctx, goalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if plan == nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "goal has no verification plan"})
		return
	}

	outcome := models.VerificationOutcomePending
	switch plan.Type {
	case models.VerificationTypeEvidence, models.VerificationTypeValue:
		evidence, err := h.verification.ListEvidence(ctx, goalID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
			return
		}
		if len(evidence) > 0 {
			outcome = models.VerificationOutcomePassed
		}
	case models.VerificationTypeSubgoal:
		passed, err := h.verification.HasPassedVerification(ctx, goalID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
			return
		}
		if passed {
			outcome = models.VerificationOutcomePassed
		}
	}

	verificationID := &plan.ID

	result, err := h.verification.CreateResult(ctx, goalID, verificationID, outcome, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ApproveVerificationResult godoc
// @Summary Approve a verification outcome
// @Description Approves a verification outcome. Only a person can do this - an agent may create outcomes but never approve them.
// @Tags goals
// @Produce json
// @Security BearerAuth
// @Param id path int true "Verification result ID"
// @Success 200 {object} models.GoalVerificationResult
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/verification-results/{id}/approve [post]
func (h *GoalVerificationHandler) ApproveVerificationResult(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	resultID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid result ID"})
		return
	}

	result, err := h.verification.ApproveResult(c.Request.Context(), resultID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if result == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "verification result not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// RejectVerificationResult godoc
// @Summary Reject a verification outcome
// @Description Rejects a verification outcome. Only a person can do this - an agent may create outcomes but never decide on them.
// @Tags goals
// @Produce json
// @Security BearerAuth
// @Param id path int true "Verification result ID"
// @Success 200 {object} models.GoalVerificationResult
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/verification-results/{id}/reject [post]
func (h *GoalVerificationHandler) RejectVerificationResult(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	resultID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid result ID"})
		return
	}

	result, err := h.verification.RejectResult(c.Request.Context(), resultID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if result == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "verification result not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// goalAccess läser user och mål ur anropet och kontrollerar behörighet.
func (h *GoalVerificationHandler) goalAccess(c *gin.Context) (int64, int64, bool) {
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return 0, 0, false
	}

	goalID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid goal ID"})
		return 0, 0, false
	}

	hasAccess, _ := h.goals.UserHasAccess(c.Request.Context(), userID, goalID)
	if !hasAccess {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Access denied"})
		return 0, 0, false
	}

	return userID, goalID, true
}

// validVerificationType svarar på om typen finns.
func validVerificationType(typ string) bool {
	switch typ {
	case models.VerificationTypeChecklist,
		models.VerificationTypeSubgoal,
		models.VerificationTypeCalendar,
		models.VerificationTypeEvidence,
		models.VerificationTypeValue,
		models.VerificationTypeCode:
		return true
	}
	return false
}

// checkDb svarar på om databasen är uppkopplad.
func (h *GoalVerificationHandler) checkDb(c *gin.Context) bool {
	if !h.dbConnected {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Message: "Database not connected"})
		return false
	}
	return true
}
