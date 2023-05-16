package user_handles

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	tokenhandlers "userservice-go/handlers/token-handlers"
	"userservice-go/types"
)

func FindUsers(findUsersCriteria types.FindUsersCriteria) (error, []types.User) {
	var usersList []types.User
	var err error
	if len(findUsersCriteria.OrgId) == 0 && len(findUsersCriteria.Emails) == 0 && len(findUsersCriteria.UserIds) == 0 && len(findUsersCriteria.Usernames) == 0 {
		err, usersList = findAllUsers()
	} else if len(findUsersCriteria.OrgId) != 0 && len(findUsersCriteria.Emails) == 0 && len(findUsersCriteria.UserIds) == 0 && len(findUsersCriteria.Usernames) == 0 {
		err, usersList = findUsersByOrgId(findUsersCriteria)
	} else if len(findUsersCriteria.Emails) > 0 {
		err, usersList = findUsersByEmails(findUsersCriteria)
	} else if len(findUsersCriteria.Usernames) > 0 {
		err, usersList = findUsersByUserNames(findUsersCriteria)
	} else if len(findUsersCriteria.UserIds) > 0 {
		err, usersList = findUsersByUserIds(findUsersCriteria)
	}

	if err != nil {
		log.Error().Msg(err.Error())
		return err, usersList
	}

	usersList = limitResults(findUsersCriteria, usersList)

	return nil, usersList
}

func limitResults(findUsersCriteria types.FindUsersCriteria, usersList []types.User) []types.User {
	if findUsersCriteria.QueryLimit > 0 && len(usersList) > findUsersCriteria.QueryLimit {
		return append(usersList[:findUsersCriteria.QueryLimit])
	} else {
		return usersList
	}
}

func findAllUsers() (error, []types.User) {
	var usersList []types.User

	url := types.KEYCLOAK_BACKEND_URL + types.KEYCLOAK_GET_BY_USERS
	log.Info().Msg(url)

	err, users := executeGetUserHttpRequest(url)
	if err != nil {
		log.Error().Msg(err.Error())
		return err, usersList
	}
	usersList = append(usersList, users...)

	return nil, usersList
}

func findUsersByOrgId(findUsersCriteria types.FindUsersCriteria) (error, []types.User) {
	var usersList []types.User

	qPart := "q=org_id:" + findUsersCriteria.OrgId
	url := types.KEYCLOAK_BACKEND_URL + types.KEYCLOAK_GET_BY_USERS + "?" + qPart

	err, users := executeGetUserHttpRequest(url)
	if err != nil {
		log.Error().Msg(err.Error())
		return err, usersList
	}
	usersList = append(usersList, users...)
	return nil, usersList
}

func findUsersByEmails(findUsersCriteria types.FindUsersCriteria) (error, []types.User) {
	var usersList []types.User
	hostPath := types.KEYCLOAK_BACKEND_URL + types.KEYCLOAK_GET_BY_USERS
	url, _ := url.Parse(hostPath)
	queryParams := url.Query()
	if len(findUsersCriteria.OrgId) > 0 {
		queryParams.Set("q", "org_id:"+findUsersCriteria.OrgId)
	}

	for _, email := range findUsersCriteria.Emails {
		if len(email) > 0 {
			queryParams.Set("email", email)
			url.RawQuery = queryParams.Encode()
			log.Info().Msg(url.String())
			err, users := executeGetUserHttpRequest(url.String())
			if err != nil {
				log.Error().Msg(err.Error())
				return err, usersList
			}
			usersList = append(usersList, users...)
		}
	}
	return nil, usersList
}

func findUsersByUserNames(findUsersCriteria types.FindUsersCriteria) (error, []types.User) {
	var usersList []types.User
	hostPath := types.KEYCLOAK_BACKEND_URL + types.KEYCLOAK_GET_BY_USERS
	url, _ := url.Parse(hostPath)
	queryParams := url.Query()
	if len(findUsersCriteria.OrgId) > 0 {
		queryParams.Set("q", "org_id:"+findUsersCriteria.OrgId)
	}

	for _, userName := range findUsersCriteria.Usernames {
		if len(userName) != 0 {
			queryParams.Set("username", userName)
			url.RawQuery = queryParams.Encode()

			log.Info().Msg(url.String())
			err, users := executeGetUserHttpRequest(url.String())
			if err != nil {
				log.Error().Msg(err.Error())
				return err, usersList
			}
			usersList = append(usersList, users...)
		}
	}
	return nil, usersList
}

func findUsersByUserIds(findUsersCriteria types.FindUsersCriteria) (error, []types.User) {
	var usersList []types.User
	hostPath := types.KEYCLOAK_BACKEND_URL + types.KEYCLOAK_GET_BY_USERS
	url, _ := url.Parse(hostPath)
	queryParams := url.Query()
	if len(findUsersCriteria.OrgId) > 0 {
		queryParams.Set("q", "org_id:"+findUsersCriteria.OrgId)
	}

	for _, userId := range findUsersCriteria.UserIds {
		if len(userId) != 0 {
			queryParams.Set("id", userId)
			url.RawQuery = queryParams.Encode()
			log.Info().Msg(url.String())
			err, users := executeGetUserHttpRequest(url.String())
			if err != nil {
				log.Error().Msg(err.Error())
				return err, usersList
			}
			usersList = append(usersList, users...)
		}
	}
	return nil, usersList
}

func executeGetUserHttpRequest(url string) (error, []types.User) {
	var users []types.User

	err, req, client := tokenhandlers.GetHttpClientAndRequestWithToken(http.MethodGet, url, nil)
	if err != nil {
		log.Error().Msg(err.Error())
		return err, users
	}

	if client != nil && req != nil {
		response, err := client.Do(req)
		if err != nil {
			log.Error().Msg(err.Error())
			return err, users
		}

		if response.StatusCode == http.StatusOK {
			responseData, err := ioutil.ReadAll(response.Body)

			if err != nil {
				log.Error().Msg(err.Error())
				return err, users
			}
			err = json.Unmarshal(responseData, &users)
			if err != nil {
				log.Error().Msg(err.Error())
				return err, users
			}
			users = processUsersCustomAttributes(users)
		}
	}
	return nil, users
}

func processUsersCustomAttributes(users []types.User) []types.User {
	for i, user := range users {
		users[i] = processUserCustomAttributes(user)
	}

	return users
}

func processUserCustomAttributes(user types.User) types.User {
	if len(user.Attributes["is_internal"]) > 0 {
		isInternal := user.Attributes["is_internal"]
		user.IsInternal, _ = strconv.ParseBool(isInternal[0])
	}

	if len(user.Attributes["org_admin"]) > 0 {
		orgAdmin := user.Attributes["org_admin"]
		user.OrgAdmin, _ = strconv.ParseBool(orgAdmin[0])
	}

	if len(user.Attributes["type"]) > 0 {
		userType := user.Attributes["type"]
		user.Type_ = userType[0]
	}

	return user
}
