package models

// CloudService represents a self-hosted service displayed on the cloud landing page.
// To add or remove a service, edit the slice returned by CloudServices().
type CloudService struct {
	Name        string
	Description string
	URL         string
	IconURL     string
}

// CloudServices returns the list of cloud services to display.
func CloudServices() []CloudService {
	return []CloudService{
		{
			Name:        "FoundryVTT",
			Description: "Virtual tabletop for playing tabletop RPGs online",
			URL:         "https://foundry.robswebhub.net",
			IconURL:     "https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons/png/foundry-vtt.png",
		},
		{
			Name:        "Vaultwarden",
			Description: "Lightweight, self-hosted password manager compatible with Bitwarden",
			URL:         "https://vault.robswebhub.net",
			IconURL:     "https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons/svg/vaultwarden.svg",
		},
		{
			Name:        "Wakapi",
			Description: "Coding activity dashboard — tracks time spent in your editor",
			URL:         "https://wakapi.robswebhub.net",
			IconURL:     "https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons/svg/wakapi.svg",
		},
		{
			Name:        "Nextcloud",
			Description: "Personal cloud storage for files, calendars, and contacts",
			URL:         "https://storage.robswebhub.net",
			IconURL:     "https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons/svg/nextcloud.svg",
		},
		{
			Name:        "Uptime Kuma",
			Description: "Self-hosted monitoring tool to track service uptime and availability",
			URL:         "https://uptime.robswebhub.net",
			IconURL:     "https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons/svg/uptime-kuma.svg",
		},
		{
			Name:        "Audiobookshelf",
			Description: "Self-hosted audiobook and podcast server",
			URL:         "https://audiobookshelf.robswebhub.net",
			IconURL:     "https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons/svg/audiobookshelf.svg",
		},
		{
			Name:        "Paperless-ngx",
			Description: "Document management system that turns physical documents into a searchable archive of PDFs",
			URL:         "https://paperless.robswebhub.net",
			IconURL:     "https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons/svg/paperless-ngx.svg",
		},
		{
			Name:        "Miniflux",
			Description: "Minimalist and opinionated feed reader",
			URL:         "https://miniflux.robswebhub.net",
			IconURL:     "https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons/svg/miniflux.svg",
		},
	}
}
