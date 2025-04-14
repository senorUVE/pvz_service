package models

import "fmt"

type Type string

const (
	TypeElectronics Type = "электроника"
	TypeClothes     Type = "одежда"
	TypeShoes       Type = "обувь"
)

func (Type) Parse(str string) (Type, error) {
	switch str {
	case string(TypeElectronics):
		return TypeElectronics, nil
	case string(TypeClothes):
		return TypeClothes, nil
	case string(TypeShoes):
		return TypeShoes, nil
	}
	return "", fmt.Errorf("invalid type: %s", str)
}

func (t Type) String() string {
	return string(t)
}
