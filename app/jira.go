package app

import (
	"Planning_poker/app/models"
	"Planning_poker/app/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"io"
	"net/http"
	"strconv"
)

func GetTask(c *gin.Context) {
	taskKey := c.Param("taskKey")
	if taskKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'taskKey' query parameter"})
		return
	}

	jiraURL := fmt.Sprintf("%s/rest/api/2/issue/%s?fields=id", env.JiraUrl, taskKey)

	req, err := http.NewRequest("GET", jiraURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+env.JiraAPIToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to Jira"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "Jira API error"})
		return
	}

	var task map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Jira response"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func GetTaskDetails(c *gin.Context) {
	taskKey := c.Param("taskKey")
	if taskKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'taskKey' query parameter"})
		return
	}

	jiraURL := fmt.Sprintf("%s/rest/api/2/issue/%s?fields=summary,description,comment,attachment", env.JiraUrl, taskKey)

	req, err := http.NewRequest("GET", jiraURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+env.JiraAPIToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to Jira"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "Jira API error"})
		return
	}

	var response struct {
		Fields struct {
			Summary     string `json:"summary"`
			Description string `json:"description"`
			Comment     struct {
				Comments []struct {
					Author struct {
						DisplayName string `json:"displayName"`
					} `json:"author"`
					Body string `json:"body"`
				} `json:"comments"`
			} `json:"comment"`
			Attachment []struct {
				Self     string `json:"self"`
				ID       string `json:"id"`
				Filename string `json:"filename"`
				Author   struct {
					Self         string            `json:"self"`
					Name         string            `json:"name"`
					Key          string            `json:"key"`
					EmailAddress string            `json:"emailAddress"`
					AvatarUrls   map[string]string `json:"avatarUrls"`
					DisplayName  string            `json:"displayName"`
					Active       bool              `json:"active"`
					TimeZone     string            `json:"timeZone"`
				} `json:"author"`
				Created   string `json:"created"`
				Size      int    `json:"size"`
				MimeType  string `json:"mimeType"`
				Content   string `json:"content"`
				Thumbnail string `json:"thumbnail"`
			}
		} `json:"fields"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Jira response"})
		return
	}

	attachments := make(map[string]models.AttachmentInfo)
	for _, attachment := range response.Fields.Attachment {
		attachments[attachment.Filename] = models.AttachmentInfo{
			Content:  attachment.Content,
			MimeType: attachment.MimeType,
		}
	}

	filteredComments := []map[string]interface{}{}
	for _, comment := range response.Fields.Comment.Comments {
		if comment.Author.DisplayName != "gitlab_wakacje" {

			processedComment := utils.Format(comment.Body, attachments)
			filteredComments = append(filteredComments, map[string]interface{}{
				"author": comment.Author.DisplayName,
				"body":   template.HTML(processedComment),
			})
		}
	}

	description := utils.Format(response.Fields.Description, attachments)
	result := gin.H{
		"summary":     response.Fields.Summary,
		"description": template.HTML(description),
		"comments":    filteredComments,
	}

	c.JSON(http.StatusOK, result)
}

func SaveTask(c *gin.Context) {
	type RequestBody struct {
		Task string `json:"task"`
		Fib  string `json:"fib"`
	}

	var requestBody RequestBody
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if requestBody.Task == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'task'"})
		return
	}

	if requestBody.Fib == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'fib'"})
		return
	}

	jiraURL := fmt.Sprintf("%s/rest/api/2/issue/%s", env.JiraUrl, requestBody.Task)

	i, err := strconv.Atoi(requestBody.Fib)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid 'fib' value"})
		return
	}
	body := map[string]interface{}{
		"fields": map[string]interface{}{
			"customfield_10082": i,
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode request body"})
		return
	}

	req, err := http.NewRequest("PUT", jiraURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+env.JiraAPIToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to Jira"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		c.JSON(resp.StatusCode, gin.H{"error": "Jira API error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully"})
}

func ImageProxyHandler(w http.ResponseWriter, r *http.Request) {
	imageUrl := r.URL.Query().Get("url")
	if imageUrl == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	req, err := http.NewRequest("GET", imageUrl, nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", "Bearer "+env.JiraAPIToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to fetch image", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	for key, value := range resp.Header {
		w.Header()[key] = value
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
