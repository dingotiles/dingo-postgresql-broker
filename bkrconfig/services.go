package bkrconfig

type Service struct {
	ID          string          `yaml:"id"`
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Bindable    bool            `yaml:"bindable"`
	Plans       []ServicePlan   `yaml:"plans"`
	Metadata    ServiceMetadata `yaml:"metadata"`
	Tags        []string        `yaml:"tags"`
}

type ServicePlan struct {
	ID          string              `yaml:"id"`
	Name        string              `yaml:"name"`
	Description string              `yaml:"description"`
	Metadata    ServicePlanMetadata `yaml:"metadata"`
}

type ServicePlanMetadata struct {
	Bullets     []string `yaml:"bullets"`
	DisplayName string   `yaml:"displayName"`
}

type ServiceMetadata struct {
	DisplayName      string                  `yaml:"displayName"`
	LongDescription  string                  `yaml:"longDescription"`
	DocumentationUrl string                  `yaml:"documentationUrl"`
	SupportUrl       string                  `yaml:"supportUrl"`
	Listing          ServiceMetadataListing  `yaml:"listing"`
	Provider         ServiceMetadataProvider `yaml:"provider"`
}

type ServiceMetadataListing struct {
	Blurb    string `yaml:"blurb"`
	ImageUrl string `yaml:"imageUrl"`
}

type ServiceMetadataProvider struct {
	Name string `yaml:"name"`
}
