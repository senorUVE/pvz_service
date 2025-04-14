package models

import "fmt"

type City string

const (
	CityMoscow City = "Москва"
	CitySPB    City = "Санкт-Петербург"
	CityKazan  City = "Казань"
)

func (City) Parse(str string) (City, error) {
	switch str {
	case string(CityMoscow):
		return CityMoscow, nil
	case string(CitySPB):
		return CitySPB, nil
	case string(CityKazan):
		return CityKazan, nil
	}
	return "", fmt.Errorf("invalid type: %s", str)
}

func (c City) String() string {
	return string(c)
}
