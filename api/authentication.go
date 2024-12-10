package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
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

func generateJWTToken(user CmuOAuthBasicInfoDTO, notAdmin bool) (string, error) {
	firstName := user.FirstnameTH
	if firstName == "" {
		firstName = helpers.Capitalize(user.FirstnameEN)
	}
	lastName := user.LastnameTH
	if lastName == "" {
		lastName = helpers.Capitalize(user.LastnameEN)
	}
	claims := jwt.MapClaims{
		"email":     user.CmuitAccount,
		"firstName": firstName,
		"lastName":  lastName,
		"faculty":   user.OrganizationNameTH,
	}

	if user.StudentID != "" && notAdmin {
		claims["studentId"] = user.StudentID
	}

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
		var body LoginDTO
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
		err = row.Scan(&user.ID, &user.Firstname, &user.Lastname, &user.Email, &user.RoomID)
		if err == sql.ErrNoRows {
			if basicInfo.ItAccountTypeID == STUDENT.String() {
				tokenString, err := generateJWTToken(*basicInfo, true)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate JWT token"})
				}
				return c.JSON(http.StatusOK, map[string]interface{}{
					"token": tokenString,
				})
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

		if user.Firstname == nil || user.Lastname == nil {
			updateQuery := `UPDATE users SET firstname = $1, lastname = $2 WHERE email = $3`
			_, err := dbConn.Exec(updateQuery, basicInfo.FirstnameTH, basicInfo.LastnameTH, basicInfo.CmuitAccount)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user data"})
			}
			user.Firstname = &basicInfo.FirstnameTH
			user.Lastname = &basicInfo.LastnameTH
		}

		return c.JSON(http.StatusOK, helpers.FormatSuccessResponse(map[string]interface{}{
			"token": tokenString,
			"user":  user,
		}))
	}
}
