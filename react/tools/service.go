/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tools

import (
	"context"
	"fmt"
	"strings"
)

// fake service 模拟的后端服务的 service
// 提供 QueryDishes, QueryRestaurants 两个方法.
var restService = &fakeService{
	repo: database,
}

// fake database.
var database = &restaurantDatabase{
	restaurantByID:        make(map[string]restaurantDataItem),
	restaurantsByLocation: make(map[string][]restaurantDataItem),
}

func init() {
	// prepare database
	restData := getData()
	for location, rests := range restData {
		for _, rest := range rests {
			database.restaurantByID[rest.ID] = rest
			database.restaurantsByLocation[location] = append(database.restaurantsByLocation[location], rest)
		}
	}
}

// ====== fake service ======
type fakeService struct {
	repo *restaurantDatabase
}

// QueryRestaurants 查询一个 location 的餐厅列表.
func (ft *fakeService) QueryRestaurants(ctx context.Context, in *QueryRestaurantsParam) (out []Restaurant, err error) {
	rests, err := ft.repo.GetRestaurantsByLocation(ctx, in.Location, in.Topn)
	if err != nil {
		return nil, err
	}

	res := make([]Restaurant, 0, len(rests))
	for _, rest := range rests {

		res = append(res, Restaurant{
			ID:    rest.ID,
			Name:  rest.Name,
			Place: rest.Place,
			Score: rest.Score,
		})
	}

	return res, nil
}

// QueryDishes 根据餐厅的 id, 查询餐厅的菜品列表.
func (ft *fakeService) QueryDishes(ctx context.Context, in *QueryDishesParam) (res []Dish, err error) {
	dishes, err := ft.repo.GetDishesByRestaurant(ctx, in.RestaurantID, in.Topn)
	if err != nil {
		return nil, err
	}

	res = make([]Dish, 0, len(dishes))
	for _, dish := range dishes {
		res = append(res, Dish{
			Name:  dish.Name,
			Desc:  dish.Desc,
			Price: dish.Price,
			Score: dish.Score,
		})
	}

	return res, nil
}

type restaurantDishDataItem struct {
	Name  string `json:"name"`
	Desc  string `json:"desc"`
	Price int    `json:"price"`
	Score int    `json:"score"`
}

type restaurantDataItem struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Desc  string `json:"desc"`
	Place string `json:"place"`
	Score int    `json:"score"` // 0 - 10

	Dishes []restaurantDishDataItem `json:"dishes"` // 餐厅中的菜
}

type restaurantDatabase struct {
	restaurantByID        map[string]restaurantDataItem   // id => restaurantDataItem
	restaurantsByLocation map[string][]restaurantDataItem // location => []restaurantDataItem
}

func (rd *restaurantDatabase) GetRestaurantsByLocation(ctx context.Context, location string, topn int) ([]restaurantDataItem, error) {
	for locationName, rests := range rd.restaurantsByLocation {
		if strings.Contains(locationName, location) || strings.Contains(location, locationName) {

			res := make([]restaurantDataItem, 0, len(rests))
			for i := 0; i < topn && i < len(rests); i++ {
				res = append(res, rests[i])
			}

			return res, nil
		}
	}

	return nil, fmt.Errorf("location %s not found", location)
}

func (rd *restaurantDatabase) GetDishesByRestaurant(ctx context.Context, restaurantID string, topn int) ([]restaurantDishDataItem, error) {
	rest, ok := rd.restaurantByID[restaurantID]
	if !ok {
		return nil, fmt.Errorf("restaurant %s not found", restaurantID)
	}

	res := make([]restaurantDishDataItem, 0, len(rest.Dishes))

	for i := 0; i < topn && i < len(rest.Dishes); i++ {
		res = append(res, rest.Dishes[i])
	}

	return res, nil
}

func getData() map[string][]restaurantDataItem { // nolint: byted_s_too_many_lines_in_func
	return map[string][]restaurantDataItem{
		"Beijing": {
			{
				ID:    "1001",
				Name:  "Cloud Edge Restaurant",
				Place: "Beijing",
				Desc:  "This is Cloud Edge Restaurant in Beijing, with diverse flavors",
				Score: 3,
				Dishes: []restaurantDishDataItem{
					{
						Name:  "Braised Pork",
						Desc:  "A piece of braised pork",
						Price: 20,
						Score: 8,
					},
					{
						Name:  "Spring Water Beef",
						Desc:  "Lots of boiled beef",
						Price: 50,
						Score: 8,
					},
					{
						Name:  "Stir-fried Pumpkin",
						Desc:  "Mushy stir-fried pumpkin",
						Price: 5,
						Score: 5,
					},
					{
						Name:  "Korean Spicy Cabbage",
						Desc:  "This is blessed spicy cabbage, very delicious",
						Price: 20,
						Score: 9,
					},
					{
						Name:  "Hot and Sour Shredded Potatoes",
						Desc:  "Hot and sour shredded potatoes",
						Price: 10,
						Score: 9,
					},
					{
						Name:  "Hot and Sour Noodles",
						Desc:  "Hot and sour noodles",
						Price: 5,
					},
				},
			},
			{
				ID:    "1002",
				Name:  "Jufu Mansion Restaurant",
				Place: "Beijing",
				Desc:  "Jufu Mansion Restaurant in Beijing, many food stalls waiting for you to explore",
				Score: 5,
				Dishes: []restaurantDishDataItem{
					{
						Name:  "Braised Spare Ribs",
						Desc:  "Pieces of spare ribs",
						Price: 43,
						Score: 7,
					},
					{
						Name:  "Big Knife Twice-Cooked Pork",
						Desc:  "Classic twice-cooked pork, large pieces",
						Price: 40,
						Score: 8,
					},
					{
						Name:  "Fiery Kiss",
						Desc:  "Cold pig snout, spicy but not greasy",
						Price: 60,
						Score: 9,
					},
					{
						Name:  "Spicy Preserved Egg",
						Desc:  "Ground chili with preserved egg, perfect with rice",
						Price: 15,
						Score: 8,
					},
				},
			},
			{
				ID:    "1003",
				Name:  "Flower Shadow Restaurant",
				Place: "Shanghai",
				Desc:  "Very luxurious Flower Shadow Restaurant, delicious and affordable",
				Score: 10,
				Dishes: []restaurantDishDataItem{
					{
						Name:  "Super Braised Pork",
						Desc:  "A very glossy piece of braised pork",
						Price: 30,
						Score: 9,
					},
					{
						Name:  "Super Beijing Roast Duck",
						Desc:  "Rolled roast duck with sauce",
						Price: 60,
						Score: 9,
					},
					{
						Name:  "Super Chinese Cabbage",
						Desc:  "Just watery stir-fried Chinese cabbage",
						Price: 8,
						Score: 8,
					},
				},
			},
		},
		"Shanghai": {
			{
				ID:    "2001",
				Name:  "Hongbin Elegant Restaurant",
				Place: "Shanghai",
				Desc:  "This is Hongbin Elegant Restaurant in Shanghai, with diverse flavors",
				Score: 3,
				Dishes: []restaurantDishDataItem{
					{
						Name:  "Sweet and Sour Tomatoes",
						Desc:  "Sweet and sour tomatoes",
						Price: 80,
						Score: 5,
					},
					{
						Name:  "Candied Fish",
						Desc:  "Fish with lots of sugar, as famous as vinegar fish",
						Price: 99,
						Score: 6,
					},
				},
			},
			{
				ID:    "2002",
				Name:  "Food Drunk Gang Base",
				Desc:  "Focused on sweet and sour flavors, worth having",
				Place: "Shanghai",
				Score: 5,
				Dishes: []restaurantDishDataItem{
					{
						Name:  "Sweet and Sour Watermelon",
						Desc:  "Sweet and sour flavor, crispy",
						Price: 69,
						Score: 7,
					},
					{
						Name:  "Sweet and Sour Big Bun",
						Desc:  "As famous as Tianjin Goubuli",
						Price: 99,
						Score: 4,
					},
				},
			},
			{
				ID:    "2010",
				Name:  "So Good You'll Stamp Your Feet Restaurant",
				Desc:  "This is the So Good You'll Stamp Your Feet Restaurant, hidden in a place you can't find, waiting for destined customers to explore. Mainly Sichuan cuisine, with generous amounts of chili and Sichuan pepper.",
				Place: "It's where it isn't",
				Score: 10,
				Dishes: []restaurantDishDataItem{
					{
						Name:  "Unbeatable Spicy Shrimp",
						Desc:  "Extremely fragrant and aromatic",
						Price: 199,
						Score: 9,
					},
					{
						Name:  "Super Hot Pot",
						Desc:  "Hot pot with lots of chili and rice wine, for cooking things like apples and bananas",
						Price: 198,
						Score: 9,
					},
				},
			},
		},
	}
}
