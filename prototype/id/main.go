package main

import (
	"fmt"

	"github.com/ketMix/ebijam25/internal/world"
)

func main() {
	var id world.SchlubID
	fmt.Println("fam", id.FamilyID())
	fmt.Println("schl", id.SchlubID())
	fmt.Println("kind", id.KindID())
	nextSchlub := id.NextSchlub().NextSchlub()
	nextSchlub.SetKindID(int(world.SchlubKindMonk))
	fmt.Println("fam", nextSchlub.FamilyID())
	fmt.Println("schl", nextSchlub.SchlubID())
	fmt.Println("kind", nextSchlub.KindID())
	nextFamily := nextSchlub.NextFamily().NextFamily().NextFamily()
	fmt.Println("fam", nextFamily.FamilyID())
	fmt.Println("schl", nextFamily.SchlubID())
	fmt.Println("kind", nextFamily.KindID())

	id = nextFamily.NextSchlub()
	id.SetKindID(4)
	id.SetItemID(14)
	fmt.Println("das schlubbe", id)
}
