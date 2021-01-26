package timerjob

import (
	"log"

	"github.com/sporule/grater/models"
)

//GenerateLinks refresh the links for the given rule
func GenerateLinks(rule models.Rule) {
	rule.GenerateAndInsertLinks()
	log.Println("Generated links for rule:", rule.Name)
}
