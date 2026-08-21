package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/kyc/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/kyc/domain/entities"
)

type KYCHandler struct {
	listPendingUC    *usecases.ListPendingKYC
	getDocumentsUC   *usecases.GetDriverDocuments
	reviewDocumentUC *usecases.ReviewDocument
}

func NewKYCHandler(
	listPendingUC *usecases.ListPendingKYC,
	getDocumentsUC *usecases.GetDriverDocuments,
	reviewDocumentUC *usecases.ReviewDocument,
) *KYCHandler {
	return &KYCHandler{
		listPendingUC:    listPendingUC,
		getDocumentsUC:   getDocumentsUC,
		reviewDocumentUC: reviewDocumentUC,
	}
}

func (h *KYCHandler) ListPendingKYC(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	drivers, err := h.listPendingUC.Execute(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_pending_kyc"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"drivers": drivers})
}

func (h *KYCHandler) GetDriverDocuments(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_user_id"})
		return
	}

	docs, err := h.getDocumentsUC.Execute(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_documents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"documents": docs})
}

type ReviewDocumentRequest struct {
	Status string  `json:"status" binding:"required"` // APPROVED or REJECTED
	Reason *string `json:"reason"`
}

func (h *KYCHandler) ReviewDocument(c *gin.Context) {
	docID, err := uuid.Parse(c.Param("documentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_document_id"})
		return
	}

	var req ReviewDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	err = h.reviewDocumentUC.Execute(c.Request.Context(), usecases.ReviewDocumentInput{
		DocumentID: docID,
		Status:     entities.DocumentStatus(req.Status),
		Reason:     req.Reason,
	})
	if err != nil {
		if err == entities.ErrDocumentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "document_not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_review_document"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "document_reviewed"})
}
