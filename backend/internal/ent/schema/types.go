package schema

import "fmt"

// UnitValues defines the standard measurement units commonly used.
var UnitValues = []string{
	"g",     // grams
	"kg",    // kilograms
	"ml",    // milliliters
	"l",     // liters
	"tsp",   // teaspoon
	"tbsp",  // tablespoon
	"cup",   // cup
	"oz",    // ounce
	"lb",    // pound
	"pc",    // piece/count
	"pinch", // pinch
}

// RegionValues defines geographic/cuisine regions for recipes and ingredients.
var RegionValues = []string{
	"asia",
	"europe",
	"north_america",
	"south_america",
	"africa",
	"oceania",
	"middle_east",
}

// MealTypeValues defines the type of a meal a recipe is suited for.
var MealTypeValues = []string{
	"breakfast",
	"lunch",
	"dinner",
	"snack",
	"dessert",
	"appetizer",
	"side_dish",
	"beverage",
	"experimental",
}

// DifficultyValues defines recipe difficulty levels.
var DifficultyValues = []string{
	"easy",
	"medium",
	"hard",
	"expert",
	"grandma",
}

// CuisineValues narrows down the type of cuisine a recipe belongs to.
var CuisineValues = []string{
	"italian",
	"french",
	"mexican",
	"chinese",
	"japanese",
	"indian",
	"thai",
	"vietnamese",
	"korean",
	"greek",
	"spanish",
	"american",
	"british",
	"german",
	"moroccan",
	"ethiopian",
	"brazilian",
	"peruvian",
	"turkish",
	"lebanese",
	"australian",
	"caribbean",
	"fusion",
	"other",
}

// AllergenValues defines the EU major allergens.
// Based on the EU regulation: https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX:32011R1169
var AllergenValues = []string{
	"gluten",      // Cereals containing gluten (wheat, rye, barley, oats, spelt, kamut)
	"crustaceans", // Crustaceans and products thereof
	"eggs",        // Eggs and products thereof
	"fish",        // Fish and products thereof
	"peanuts",     // Peanuts and products thereof
	"soybeans",    // Soybeans and products thereof
	"milk",        // Milk and products thereof (including lactose)
	"nuts",        // Tree nuts (almonds, hazelnuts, walnuts, cashews, pecans, Brazil nuts, pistachios, macadamia)
	"celery",      // Celery and products thereof
	"mustard",     // Mustard and products thereof
	"sesame",      // Sesame seeds and products thereof
	"sulphites",   // Sulphur dioxide and sulphites
	"lupin",       // Lupin and products thereof
	"molluscs",    // Molluscs and products thereof
}

var allergenValueSet = valueSet(AllergenValues)

func valueSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
}

// ValidateAllergenValues enforces that recipe allergens only use known EU allergen identifiers.
func ValidateAllergenValues(values []string) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, ok := allergenValueSet[value]; !ok {
			return fmt.Errorf("invalid allergen %q", value)
		}
		if _, ok := seen[value]; ok {
			return fmt.Errorf("duplicate allergen %q", value)
		}
		seen[value] = struct{}{}
	}
	return nil
}
