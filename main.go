package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ClusterRequest struct {
	Bbox              string `json:"bbox" binding:"required"`
	Zoom              int    `json:"zoom" binding:"required"`
	PartialRedemption bool   `json:"partial_redemption"`
	ReturnOption      bool   `json:"return_option"`
	DeliveryService   string `json:"delivery_service"`
}

type ClusterResponse struct {
	Count int `json:"count"`
	Data  []struct {
		IdCluster int `json:"IdCluster"`
		Geometry  struct {
			Coordinates []string `json:"coordinates"`
			Type        string   `json:"type"`
		} `json:"geometry"`
		Properties struct {
			PointCount int  `json:"pointCount"`
			IsPremium  bool `json:"isPremium"`
		} `json:"properties"`
	} `json:"data"`
}

func main() {
	r := gin.Default()

	r.Static("/static", "./static")
	r.LoadHTMLFiles("templates/index.html")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.POST("/clusters", func(c *gin.Context) {
		var req ClusterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Используем zoom из запроса
		apiResponse, err := fetchClusters(req, req.Zoom)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, apiResponse)
	})

	r.Run(":8083")
}

func fetchClusters(req ClusterRequest, zoom int) (*ClusterResponse, error) {
	client := &http.Client{}

	// Базовый URL
	baseUrl := "https://offers.sokolov.ru/api/v1/offer/points-cluster/"

	// Формируем параметры
	queryParams := url.Values{}
	queryParams.Set("filter[bbox]", req.Bbox)
	queryParams.Set("filter[zoom]", strconv.Itoa(zoom))

	if req.PartialRedemption {
		queryParams.Set("filter[partial_redemption]", "true")
	}
	if req.ReturnOption {
		queryParams.Set("filter[return_option]", "true")
	}
	if req.DeliveryService != "" {
		queryParams.Set("filter[delivery_service]", req.DeliveryService)
	}

	queryParams.Set("filter[offer]", "92021398")
	queryParams.Set("filter[point_type]", "pvz")

	// Объединяем базовый URL с параметрами
	fullUrl := baseUrl + "?" + queryParams.Encode()
	fmt.Println(fullUrl)
	// Отправляем запрос
	resp, err := client.Get(fullUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var clusterResponse ClusterResponse
	if err := json.NewDecoder(resp.Body).Decode(&clusterResponse); err != nil {
		return nil, err
	}

	return &clusterResponse, nil
}
