//go:build unit

// Package globalvm provides global scheduling capabilities for the distributed compute network.
// This file implements geographic location detection for nodes.
package globalvm

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bacalhau-project/bacalhau/pkg/models"
)

// Location represents the geographic location of a node.
type Location struct {
	// Region is the geographic region (e.g., "us-east", "eu-west").
	Region string `json:"Region,omitempty"`

	// Zone is the availability zone (e.g., "us-east-1a").
	Zone string `json:"Zone,omitempty"`

	// Country is the country code (ISO 3166-1 alpha-2).
	Country string `json:"Country,omitempty"`

	// City is the city name.
	City string `json:"City,omitempty"`

	// Latitude is the geographic latitude.
	Latitude float64 `json:"Latitude,omitempty"`

	// Longitude is the geographic longitude.
	Longitude float64 `json:"Longitude,omitempty"`

	// ISP is the internet service provider.
	ISP string `json:"ISP,omitempty"`

	// CloudProvider is the detected cloud provider (aws, gcp, azure, etc).
	CloudProvider string `json:"CloudProvider,omitempty"`

	// CloudRegion is the cloud provider region.
	CloudRegion string `json:"CloudRegion,omitempty"`

	// Source indicates how the location was determined.
	Source string `json:"Source,omitempty"` // "metadata", "geoip", "config", "default"
}

// LocationDetector detects the geographic location of a node.
type LocationDetector interface {
	// DetectLocation detects the location of the current node.
	DetectLocation(ctx context.Context) (*Location, error)

	// DetectLocationFromIP detects location from a specific IP.
	DetectLocationFromIP(ctx context.Context, ip string) (*Location, error)

	// GetRegion returns the region for a node based on its location.
	GetRegion(loc *Location) string
}

// LocationDetectorConfig configures the location detector.
type LocationDetectorConfig struct {
	// ConfiguredRegion is a manually configured region (highest priority).
	ConfiguredRegion string

	// ConfiguredZone is a manually configured availability zone.
	ConfiguredZone string

	// UseCloudMetadata enables cloud provider metadata detection.
	UseCloudMetadata bool

	// UseGeoIP enables IP geolocation.
	UseGeoIP bool

	// GeoIPAPI is the geolocation API endpoint.
	GeoIPAPI string

	// MetadataTimeout is the timeout for metadata API calls.
	MetadataTimeout time.Duration

	// DefaultRegion is used when location cannot be determined.
	DefaultRegion string
}

// DefaultLocationDetectorConfig returns the default configuration.
func DefaultLocationDetectorConfig() LocationDetectorConfig {
	return LocationDetectorConfig{
		UseCloudMetadata: true,
		UseGeoIP:         true,
		GeoIPAPI:         "https://ipinfo.io/json",
		MetadataTimeout:  2 * time.Second,
		DefaultRegion:    "default",
	}
}

// locationDetector implements the LocationDetector interface.
type locationDetector struct {
	config LocationDetectorConfig
	client *http.Client

	mu       sync.RWMutex
	cached   *Location
	cacheTTL time.Duration
}

// NewLocationDetector creates a new location detector.
func NewLocationDetector(config LocationDetectorConfig) LocationDetector {
	return &locationDetector{
		config: config,
		client: &http.Client{
			Timeout: config.MetadataTimeout,
		},
		cacheTTL: 5 * time.Minute,
	}
}

// DetectLocation detects the location of the current node.
func (d *locationDetector) DetectLocation(ctx context.Context) (*Location, error) {
	// Check cache first
	d.mu.RLock()
	if d.cached != nil {
		defer d.mu.RUnlock()
		return d.cached, nil
	}
	d.mu.RUnlock()

	loc := &Location{
		Source: "default",
	}

	// Priority 1: Manual configuration (highest priority)
	if d.config.ConfiguredRegion != "" {
		loc.Region = d.config.ConfiguredRegion
		loc.Zone = d.config.ConfiguredZone
		loc.Source = "config"
		d.mu.Lock()
		d.cached = loc
		d.mu.Unlock()
		return loc, nil
	}

	// Priority 2: Cloud provider metadata
	if d.config.UseCloudMetadata {
		if cloudLoc := d.detectCloudMetadata(ctx); cloudLoc != nil {
			loc = cloudLoc
			loc.Source = "metadata"
		}
	}

	// Priority 3: IP geolocation (if no cloud metadata found)
	if loc.Source == "default" && d.config.UseGeoIP {
		if geoLoc, err := d.detectGeoIP(ctx); err == nil {
			loc = geoLoc
			loc.Source = "geoip"
		}
	}

	// Fallback to default region
	if loc.Region == "" {
		loc.Region = d.config.DefaultRegion
	}

	d.mu.Lock()
	d.cached = loc
	d.mu.Unlock()

	return loc, nil
}

// DetectLocationFromIP detects location from a specific IP.
func (d *locationDetector) DetectLocationFromIP(ctx context.Context, ip string) (*Location, error) {
	if d.config.GeoIPAPI == "" {
		return &Location{Region: d.config.DefaultRegion, Source: "default"}, nil
	}

	// Use ipinfo.io API for geolocation
	url := fmt.Sprintf("https://ipinfo.io/%s/json", ip)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query geoip: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Country  string `json:"country"`
		Region   string `json:"region"`
		City     string `json:"city"`
		Loc      string `json:"loc"` // "lat,lon"
		Org      string `json:"org"` // ISP
		Timezone string `json:"timezone"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode geoip response: %w", err)
	}

	loc := &Location{
		Country: result.Country,
		City:    result.City,
		ISP:     result.Org,
		Source:  "geoip",
	}

	// Parse coordinates
	if result.Loc != "" {
		parts := strings.Split(result.Loc, ",")
		if len(parts) == 2 {
			fmt.Sscanf(parts[0], "%f", &loc.Latitude)
			fmt.Sscanf(parts[1], "%f", &loc.Longitude)
		}
	}

	// Convert to region
	loc.Region = d.GetRegion(loc)

	return loc, nil
}

// GetRegion returns the region for a node based on its location.
func (d *locationDetector) GetRegion(loc *Location) string {
	// If region is already set, return it
	if loc.Region != "" {
		return loc.Region
	}

	// If cloud region is available, use it
	if loc.CloudRegion != "" {
		return loc.CloudRegion
	}

	// Map country to a region
	regionMap := map[string]string{
		"US": "us-east", // Default US to east
		"CA": "us-east",
		"BR": "south-america",
		"GB": "eu-west",
		"DE": "eu-central",
		"FR": "eu-west",
		"NL": "eu-west",
		"IE": "eu-west",
		"JP": "asia-east",
		"SG": "asia-south",
		"IN": "asia-south",
		"AU": "asia-east",
		"KR": "asia-east",
		"HK": "asia-east",
		"CN": "asia-east",
	}

	if region, ok := regionMap[loc.Country]; ok {
		return region
	}

	return d.config.DefaultRegion
}

// detectCloudMetadata attempts to detect location from cloud provider metadata.
func (d *locationDetector) detectCloudMetadata(ctx context.Context) *Location {
	// Try AWS metadata first
	if loc := d.detectAWSMetadata(ctx); loc != nil {
		return loc
	}

	// Try GCP metadata
	if loc := d.detectGCPMetadata(ctx); loc != nil {
		return loc
	}

	// Try Azure metadata
	if loc := d.detectAzureMetadata(ctx); loc != nil {
		return loc
	}

	return nil
}

// detectAWSMetadata detects location from AWS EC2 metadata.
func (d *locationDetector) detectAWSMetadata(ctx context.Context) *Location {
	// AWS IMDSv2 requires a session token
	// First, try to get a token
	tokenURL := "http://169.254.169.254/latest/api/token"
	tokenReq, err := http.NewRequestWithContext(ctx, http.MethodPut, tokenURL, nil)
	if err != nil {
		return nil
	}
	tokenReq.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", "60")

	tokenResp, err := d.client.Do(tokenReq)
	if err != nil {
		return nil // Not on AWS
	}
	defer tokenResp.Body.Close()

	var token string
	if tokenResp.StatusCode == http.StatusOK {
		var tokenBytes []byte
		if _, err := fmt.Fscanf(tokenResp.Body, "%s", &tokenBytes); err != nil {
			// IMDSv1 fallback
		} else {
			token = string(tokenBytes)
		}
	}

	// Get identity document
	identityURL := "http://169.254.169.254/latest/dynamic/instance-identity/document"
	identityReq, err := http.NewRequestWithContext(ctx, http.MethodGet, identityURL, nil)
	if err != nil {
		return nil
	}

	if token != "" {
		identityReq.Header.Set("X-aws-ec2-metadata-token", token)
	}

	identityResp, err := d.client.Do(identityReq)
	if err != nil {
		return nil
	}
	defer identityResp.Body.Close()

	if identityResp.StatusCode != http.StatusOK {
		return nil
	}

	var identity struct {
		Region      string `json:"region"`
		Zone        string `json:"availabilityZone"`
		InstanceID  string `json:"instanceId"`
		AccountID   string `json:"accountId"`
	}

	if err := json.NewDecoder(identityResp.Body).Decode(&identity); err != nil {
		return nil
	}

	loc := &Location{
		CloudProvider: "aws",
		CloudRegion:   identity.Region,
		Zone:          identity.Zone,
		Region:        identity.Region,
		Source:        "metadata",
	}

	return loc
}

// detectGCPMetadata detects location from GCE metadata.
func (d *locationDetector) detectGCPMetadata(ctx context.Context) *Location {
	// GCE metadata service
	metadataURL := "http://metadata.google.internal/computeMetadata/v1/instance/?recursive=true"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, metadataURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Metadata-Flavor", "Google")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil // Not on GCP
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var metadata struct {
		Zone string `json:"zone"`
		Name string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil
	}

	// Parse zone (projects/{project}/zones/{zone})
	zone := metadata.Zone
	if idx := strings.LastIndex(zone, "/"); idx >= 0 {
		zone = zone[idx+1:]
	}

	// Extract region from zone (e.g., us-central1-a -> us-central1)
	region := zone
	if idx := strings.LastIndex(zone, "-"); idx >= 0 {
		region = zone[:idx]
	}

	loc := &Location{
		CloudProvider: "gcp",
		CloudRegion:   region,
		Zone:          zone,
		Region:        region,
		Source:        "metadata",
	}

	return loc
}

// detectAzureMetadata detects location from Azure Instance Metadata Service.
func (d *locationDetector) detectAzureMetadata(ctx context.Context) *Location {
	// Azure IMDS
	metadataURL := "http://169.254.169.254/metadata/instance?api-version=2021-02-01"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, metadataURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Metadata", "true")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil // Not on Azure
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var metadata struct {
		Compute struct {
			Location          string `json:"location"`
			Zone              string `json:"zone"`
			VMScaleSetName    string `json:"vmScaleSetName"`
			ResourceGroupName string `json:"resourceGroupName"`
		} `json:"compute"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil
	}

	loc := &Location{
		CloudProvider: "azure",
		CloudRegion:   metadata.Compute.Location,
		Zone:          metadata.Compute.Zone,
		Region:        metadata.Compute.Location,
		Source:        "metadata",
	}

	return loc
}

// detectGeoIP detects location using IP geolocation.
func (d *locationDetector) detectGeoIP(ctx context.Context) (*Location, error) {
	if d.config.GeoIPAPI == "" {
		return nil, fmt.Errorf("geoip api not configured")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.config.GeoIPAPI, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query geoip: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		IP       string `json:"ip"`
		Country  string `json:"country"`
		Region   string `json:"region"`
		City     string `json:"city"`
		Loc      string `json:"loc"`
		Org      string `json:"org"`
		Timezone string `json:"timezone"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode geoip response: %w", err)
	}

	loc := &Location{
		Country: result.Country,
		City:    result.City,
		ISP:     result.Org,
		Source:  "geoip",
	}

	// Parse coordinates
	if result.Loc != "" {
		parts := strings.Split(result.Loc, ",")
		if len(parts) == 2 {
			fmt.Sscanf(parts[0], "%f", &loc.Latitude)
			fmt.Sscanf(parts[1], "%f", &loc.Longitude)
		}
	}

	// Set region
	loc.Region = d.GetRegion(loc)

	return loc, nil
}

// GetLocationFromEnv gets location from environment variables.
func GetLocationFromEnv() *Location {
	loc := &Location{Source: "config"}

	// Check for common environment variables
	if region := os.Getenv("BACALHAU_REGION"); region != "" {
		loc.Region = region
	}
	if zone := os.Getenv("BACALHAU_ZONE"); zone != "" {
		loc.Zone = zone
	}
	if region := os.Getenv("AWS_REGION"); region != "" {
		loc.CloudProvider = "aws"
		loc.CloudRegion = region
		loc.Region = region
	}
	if region := os.Getenv("AWS_DEFAULT_REGION"); region != "" {
		loc.CloudProvider = "aws"
		loc.CloudRegion = region
		loc.Region = region
	}
	if zone := os.Getenv("AWS_AVAILABILITY_ZONE"); zone != "" {
		loc.Zone = zone
	}
	if region := os.Getenv("GOOGLE_CLOUD_REGION"); region != "" {
		loc.CloudProvider = "gcp"
		loc.CloudRegion = region
		loc.Region = region
	}
	if zone := os.Getenv("GOOGLE_CLOUD_ZONE"); zone != "" {
		loc.Zone = zone
	}
	if region := os.Getenv("AZURE_REGION"); region != "" {
		loc.CloudProvider = "azure"
		loc.CloudRegion = region
		loc.Region = region
	}

	if loc.Region != "" {
		return loc
	}

	return nil
}

// GetLocationFromNodeInfo extracts location from node labels.
func GetLocationFromNodeInfo(info models.NodeInfo) *Location {
	loc := &Location{Source: "labels"}

	if info.Labels == nil {
		return nil
	}

	if region, ok := info.Labels["region"]; ok {
		loc.Region = region
	}
	if zone, ok := info.Labels["zone"]; ok {
		loc.Zone = zone
	}
	if country, ok := info.Labels["country"]; ok {
		loc.Country = country
	}
	if city, ok := info.Labels["city"]; ok {
		loc.City = city
	}
	if provider, ok := info.Labels["cloud-provider"]; ok {
		loc.CloudProvider = provider
	}
	if cloudRegion, ok := info.Labels["cloud-region"]; ok {
		loc.CloudRegion = cloudRegion
	}

	// Also check Kubernetes-style labels
	if region, ok := info.Labels["topology.kubernetes.io/region"]; ok {
		loc.Region = region
	}
	if zone, ok := info.Labels["topology.kubernetes.io/zone"]; ok {
		loc.Zone = zone
	}

	if loc.Region != "" {
		return loc
	}

	return nil
}

// ExtractIP extracts the IP address from a node address string.
func ExtractIP(addr string) string {
	// Handle various formats: "IP:port", "hostname:port", "IP", "hostname"
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// No port, use as-is
		host = addr
	}

	// Validate it's an IP
	if ip := net.ParseIP(host); ip != nil {
		return ip.String()
	}

	// Not an IP, could be a hostname
	return host
}

// RegionToContinent maps regions to continents for broader grouping.
func RegionToContinent(region string) string {
	regionMap := map[string]string{
		"us-east":       "north-america",
		"us-west":       "north-america",
		"us-central":    "north-america",
		"eu-west":       "europe",
		"eu-central":    "europe",
		"eu-north":      "europe",
		"asia-east":     "asia",
		"asia-south":    "asia",
		"asia-southeast": "asia",
		"south-america": "south-america",
		"africa":        "africa",
		"australia":     "oceania",
	}

	if continent, ok := regionMap[region]; ok {
		return continent
	}

	// Try to match prefix
	for prefix, continent := range regionMap {
		if strings.HasPrefix(region, prefix) {
			return continent
		}
	}

	return "unknown"
}
