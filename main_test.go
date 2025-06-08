package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	// Arange
	handler := http.HandlerFunc(mainHandle)
	city1 := "tula"     // Настоящее значение
	city2 := "testCity" // Тестовое значение
	cafeList[city2] = make([]string, 0, 112)
	for i := range 112 {
		cafeList[city2] = append(cafeList[city2], "test "+strconv.Itoa(i))
	}
	requests := map[string][]struct {
		count int // передаваемое значение count
		want  int // ожидаемое количество кафе в ответе
	}{
		city1: {
			//{-1, 0}, тут будет паника
			{0, 0},
			{1, 1},
			{2, 2},
			{100, min(len(cafeList[city1]), 100)},
		},
		city2: {
			//{-1, 0}, тут будет паника
			{3, 3},
			{5, 5},
			{28, 28},
			{110, min(len(cafeList[city2]), 110)},
		},
	}

	for city, cityRequests := range requests {
		for ind, v := range cityRequests {

			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/cafe?city=%s&count=%d", city, v.count), nil)

			// Act
			handler.ServeHTTP(response, req)

			// Assert
			require.Equal(t, http.StatusOK, response.Code, "Вариант %d. Ожидаем статус 200", ind)

			actual := 0
			result := strings.TrimSpace(response.Body.String())
			if result != "" {
				actual = len(strings.Split(result, ","))
			}
			assert.Equal(t, v.want, actual, "Вариант %d. Сравниваем кол-во", ind)
		}
	}
}

func TestCafeSearch(t *testing.T) {
	// Arange
	handler := http.HandlerFunc(mainHandle)

	city := "moscow"

	requests := []struct {
		search    string // передаваемое значение search
		wantCount int    // ожидаемое количество кафе в ответе
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
		{"КО", 3},
		{"89", 0},
	}

	for ind, v := range requests {

		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/cafe?city=%s&search=%s", city, v.search), nil)

		// Act
		handler.ServeHTTP(response, req)

		// Assert
		require.Equal(t, http.StatusOK, response.Code, "Вариант %d. Ожидаем статус 200", ind)

		actual := 0
		result := strings.TrimSpace(response.Body.String())
		if result != "" {
			actual = len(strings.Split(result, ","))
		}
		require.Equal(t, v.wantCount, actual, "Вариант %d. Сравниваем кол-во", ind)

		if actual > 0 {
			for _, cafe := range strings.Split(result, ",") {
				fmt.Println(cafe)
			}
		}
	}
}
