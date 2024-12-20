package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"src/helpers"
	"src/models"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

type CMU_OAUTH_ROLE int

const (
	MIS CMU_OAUTH_ROLE = iota
	STUDENT
	ALUMNI
	RESIGN
	MANAGER
	NON_MIS
	ORG
	PROJECT
	RETIRED
	VIP
)

func (r CMU_OAUTH_ROLE) String() string {
	return [...]string{
		"MISEmpAcc",
		"StdAcc",
		"AlumAcc",
		"EmpResiAcc",
		"ManAcc",
		"NonMISEmpAcc",
		"OrgAcc",
		"ProjAcc",
		"RetEmpAcc",
		"VIPAcc",
	}[r]
}

type LoginDTO struct {
	Topic     int     `json:"topic" validate:"required"`
	Note      *string `json:"note"`
	FirstName string  `json:"firstName" validate:"required"`
	LastName  string  `json:"lastName" validate:"required"`
}

type AuthDTO struct {
	Code        string `json:"code" validate:"required"`
	RedirectURI string `json:"redirectUri" validate:"required"`
}

type CmuOAuthBasicInfoDTO struct {
	CmuitAccountName   string `json:"cmuitaccount_name"`
	CmuitAccount       string `json:"cmuitaccount"`
	StudentID          string `json:"student_id"`
	PrenameID          string `json:"prename_id"`
	PrenameTH          string `json:"prename_TH"`
	PrenameEN          string `json:"prename_EN"`
	FirstnameTH        string `json:"firstname_TH"`
	FirstnameEN        string `json:"firstname_EN"`
	LastnameTH         string `json:"lastname_TH"`
	LastnameEN         string `json:"lastname_EN"`
	OrganizationCode   string `json:"organization_code"`
	OrganizationNameTH string `json:"organization_name_TH"`
	OrganizationNameEN string `json:"organization_name_EN"`
	ItAccountTypeID    string `json:"itaccounttype_id"`
	ItAccountTypeTH    string `json:"itaccounttype_TH"`
	ItAccountTypeEN    string `json:"itaccounttype_EN"`
}

func getOAuthAccessToken(code, redirectUri string) (string, error) {
	client := &http.Client{}
	url := os.Getenv("CMU_OAUTH_GET_TOKEN_URL")
	data := []byte(`grant_type=authorization_code&client_id=` + os.Getenv("CMU_OAUTH_CLIENT_ID") +
		`&client_secret=` + os.Getenv("CMU_OAUTH_CLIENT_SECRET") +
		`&code=` + code +
		`&redirect_uri=` + redirectUri)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to fetch access token")
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	token, ok := result["access_token"].(string)
	if !ok {
		return "", errors.New("invalid access token response")
	}

	return token, nil
}

func getCMUBasicInfo(accessToken string) (*CmuOAuthBasicInfoDTO, error) {
	client := &http.Client{}
	url := os.Getenv("CMU_OAUTH_GET_BASIC_INFO")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch CMU basic info")
	}

	var info CmuOAuthBasicInfoDTO
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return &info, nil
}

func generateJWTToken(user interface{}, notAdmin bool) (string, error) {
	var firstName, lastName string
	claims := jwt.MapClaims{}
	switch v := user.(type) {
	case CmuOAuthBasicInfoDTO:
		claims["email"] = v.CmuitAccount
		firstName = v.FirstnameTH
		lastName = v.LastnameTH
		if v.StudentID != "" && notAdmin {
			claims["studentId"] = v.StudentID
		}
		if firstName == "" {
			firstName = helpers.Capitalize(v.FirstnameEN)
		}
		if lastName == "" {
			lastName = helpers.Capitalize(v.LastnameEN)
		}
		claims["faculty"] = v.OrganizationNameTH
	case LoginDTO:
		firstName = v.FirstName
		lastName = v.LastName
	}
	claims["firstName"] = firstName
	claims["lastName"] = lastName

	// expirationTime := time.Now().Add(7 * 24 * time.Hour)
	// claims["exp"] = expirationTime.Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secretKey := os.Getenv("JWT_SECRET_KEY")
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func Authentication(dbConn *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var body AuthDTO
		if err := c.Bind(&body); err != nil || body.Code == "" || body.RedirectURI == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid authorization code or redirect URI"})
		}
		accessToken, err := getOAuthAccessToken(body.Code, body.RedirectURI)
		if err != nil || accessToken == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot get OAuth access token"})
		}
		basicInfo, err := getCMUBasicInfo(accessToken)
		if err != nil || basicInfo == nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot get CMU basic info"})
		}

		row := dbConn.QueryRow("SELECT * FROM users WHERE email = $1", basicInfo.CmuitAccount)
		var user models.User
		err = row.Scan(&user.ID, &user.FirstNameTH, &user.LastNameTH, &user.FirstNameEN, &user.LastNameEN, &user.Email, &user.CounterID)
		if err == sql.ErrNoRows {
			if basicInfo.ItAccountTypeID == STUDENT.String() {
				tokenString, err := generateJWTToken(*basicInfo, true)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate JWT token"})
				}
				return c.JSON(http.StatusOK, helpers.FormatSuccessResponse(map[string]interface{}{
					"token": tokenString,
				}))
			} else {
				return c.JSON(http.StatusForbidden, map[string]interface{}{
					"message": "Cannot access",
				})
			}
		}
		tokenString, err := generateJWTToken(*basicInfo, false)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate JWT token"})
		}

		if user.FirstNameEN == nil || user.LastNameEN == nil {
			updateQuery := `UPDATE users SET firstname_th = $1, lastname_th = $2, firstname_en = $3, lastname_en = $4 WHERE email = $5`
			_, err := dbConn.Exec(updateQuery, basicInfo.FirstnameTH, basicInfo.LastnameTH, basicInfo.FirstnameEN, basicInfo.LastnameEN, basicInfo.CmuitAccount)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user data"})
			}
			user.FirstNameTH = &basicInfo.FirstnameTH
			user.LastNameTH = &basicInfo.LastnameTH
			user.FirstNameEN = &basicInfo.FirstnameEN
			user.LastNameEN = &basicInfo.LastnameEN
		}

		return c.JSON(http.StatusOK, helpers.FormatSuccessResponse(map[string]interface{}{
			"token": tokenString,
			"user":  user,
		}))
	}
}

func ReserveNotLogin(dbConn *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var body LoginDTO
		if err := c.Bind(&body); err != nil || body.FirstName == "" || body.LastName == "" || body.Topic == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid firstname or lastname or topic"})
		}
		tokenString, err := generateJWTToken(body, true)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate JWT token"})
		}

		var topicCode string
		topicQuery := `SELECT code FROM topics WHERE id = $1`
		err = dbConn.QueryRow(topicQuery, body.Topic).Scan(&topicCode)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve topic code"})
		}

		var lastQueueNo string
		query := `SELECT no FROM queues WHERE topic_id = $1 ORDER BY created_at DESC LIMIT 1`
		err = dbConn.QueryRow(query, body.Topic).Scan(&lastQueueNo)
		if err != nil && err != sql.ErrNoRows {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve the last queue number"})
		}

		var newQueueNo string
		if lastQueueNo != "" {
			var numPart int
			_, err := fmt.Sscanf(lastQueueNo, topicCode+"%03d", &numPart)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse the last queue number"})
			}
			numPart++
			newQueueNo = fmt.Sprintf("%s%03d", topicCode, numPart)
		} else {
			newQueueNo = fmt.Sprintf("%s001", topicCode)
		}

		var note interface{}
		if body.Note == nil {
			note = nil
		} else {
			note = *body.Note
		}

		insertQuery := `INSERT INTO queues (no, firstName, lastName, topic_id, note) 
						VALUES ($1, $2, $3, $4, $5) RETURNING id`
		_, err = dbConn.Exec(insertQuery, newQueueNo, body.FirstName, body.LastName, body.Topic, note)
		if err != nil {
			log.Printf("Error inserting queue: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create queue"})
		}

		return c.JSON(http.StatusOK, helpers.FormatSuccessResponse(map[string]interface{}{
			"token": tokenString,
		}))
	}
}
