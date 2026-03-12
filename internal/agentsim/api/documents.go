package api

import (
	"context"
	"fmt"
	"net/url"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/agent"
)

// DocumentType represents a document type returned by the API.
type DocumentType struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// DocumentCategory represents a document category.
type DocumentCategory struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Document represents a document resource.
type Document struct {
	ID                 int64  `json:"id"`
	Title              string `json:"title"`
	Subject            string `json:"subject"`
	Content            string `json:"content"`
	Status             string `json:"status"`
	DocumentTypeID     int64  `json:"document_type_id"`
	DocumentTypeName   string `json:"document_type_name"`
	CategoryID         int64  `json:"category_id"`
	AuthorID           int64  `json:"author_id"`
	AuthorName         string `json:"author_name"`
	RegistrationNumber string `json:"registration_number"`
	Importance         string `json:"importance"`
}

// DocumentList represents a paginated list of documents.
type DocumentList struct {
	Documents []Document `json:"documents"`
	Total     int        `json:"total"`
}

// GetDocumentTypes retrieves available document types.
func (c *Client) GetDocumentTypes(ctx context.Context, a *agent.Agent) ([]DocumentType, error) {
	resp, err := c.Get(ctx, "/api/document-types", a)
	if err != nil {
		return nil, err
	}
	var types []DocumentType
	if err := ParseData(resp, &types); err != nil {
		return nil, err
	}
	return types, nil
}

// GetDocumentCategories retrieves available document categories.
func (c *Client) GetDocumentCategories(ctx context.Context, a *agent.Agent) ([]DocumentCategory, error) {
	resp, err := c.Get(ctx, "/api/document-categories", a)
	if err != nil {
		return nil, err
	}
	var cats []DocumentCategory
	if err := ParseData(resp, &cats); err != nil {
		return nil, err
	}
	return cats, nil
}

// CreateDocumentRequest represents a request to create a new document.
type CreateDocumentRequest struct {
	Title          string `json:"title"`
	DocumentTypeID int64  `json:"document_type_id"`
	CategoryID     int64  `json:"category_id,omitempty"`
	Subject        string `json:"subject,omitempty"`
	Content        string `json:"content"`
	RecipientID    int64  `json:"recipient_id,omitempty"`
	Importance     string `json:"importance,omitempty"`
	IsPublic       bool   `json:"is_public,omitempty"`
}

// CreateDocument creates a new document.
func (c *Client) CreateDocument(ctx context.Context, a *agent.Agent, req CreateDocumentRequest) (*Document, error) {
	resp, err := c.Post(ctx, "/api/documents", a, req)
	if err != nil {
		return nil, fmt.Errorf("create document: %w", err)
	}
	var doc Document
	if err := ParseData(resp, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

// ListDocuments retrieves a list of documents.
func (c *Client) ListDocuments(ctx context.Context, a *agent.Agent, queryParams string) (*DocumentList, error) {
	path := "/api/documents"
	if queryParams != "" {
		path += "?" + queryParams
	}
	resp, err := c.Get(ctx, path, a)
	if err != nil {
		return nil, err
	}
	var list DocumentList
	if err := ParseData(resp, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// GetDocument retrieves a document by ID.
func (c *Client) GetDocument(ctx context.Context, a *agent.Agent, id int64) (*Document, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/documents/%d", id), a)
	if err != nil {
		return nil, err
	}
	var doc Document
	if err := ParseData(resp, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

// UpdateDocumentRequest represents a request to update a document.
type UpdateDocumentRequest struct {
	Title      string `json:"title,omitempty"`
	Subject    string `json:"subject,omitempty"`
	Content    string `json:"content,omitempty"`
	Importance string `json:"importance,omitempty"`
	IsPublic   *bool  `json:"is_public,omitempty"`
}

// UpdateDocument updates a document.
func (c *Client) UpdateDocument(ctx context.Context, a *agent.Agent, id int64, req UpdateDocumentRequest) error {
	_, err := c.Put(ctx, fmt.Sprintf("/api/documents/%d", id), a, req)
	return err
}

// ShareDocument shares a document with another user.
func (c *Client) ShareDocument(ctx context.Context, a *agent.Agent, docID int64, userID int64, permission string) error {
	body := map[string]any{
		"user_id":    userID,
		"permission": permission,
	}
	_, err := c.Post(ctx, fmt.Sprintf("/api/documents/%d/share", docID), a, body)
	return err
}

// SearchDocuments searches for documents.
func (c *Client) SearchDocuments(ctx context.Context, a *agent.Agent, query string) (*DocumentList, error) {
	path := fmt.Sprintf("/api/documents/search?q=%s", url.QueryEscape(query))
	resp, err := c.Get(ctx, path, a)
	if err != nil {
		return nil, err
	}
	var list DocumentList
	if err := ParseData(resp, &list); err != nil {
		return nil, err
	}
	return &list, nil
}
