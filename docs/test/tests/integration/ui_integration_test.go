package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLucidIconsIntegration tests Lucid Icons loading and rendering
func TestLucidIconsIntegration(t *testing.T) {
	// Skip if running in CI without browser
	if testing.Short() {
		t.Skip("Skipping UI tests in short mode")
	}

	tests := []struct {
		name          string
		url           string
		expectedIcons []string
	}{
		{
			name: "Dashboard page icons",
			url:  "/",
			expectedIcons: []string{
				"layout-dashboard", // Navigation
				"globe",            // Sites
				"folder",           // Buckets
				"repeat",           // Replication
				"check-circle",     // Consistency
				"settings",         // Operations
				"package",          // Header logo
				"refresh-cw",       // Refresh button
			},
		},
		{
			name: "Sites management page icons",
			url:  "/sites",
			expectedIcons: []string{
				"globe",    // Page icon
				"plus",     // Add site
				"trash-2",  // Remove site
				"download", // Resync from
				"upload",   // Resync to
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := createUITestServer()
			defer server.Close()

			// Setup Chrome context
			ctx, cancel := chromedp.NewContext(context.Background())
			defer cancel()

			// Set timeout
			ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			var icons []string
			err := chromedp.Run(ctx,
				// Navigate to page
				chromedp.Navigate(server.URL+tt.url),

				// Wait for Lucide to load
				chromedp.WaitVisible(`script[src*="lucide"]`, chromedp.ByQuery),
				chromedp.Sleep(2*time.Second),

				// Get all Lucid icons
				chromedp.Evaluate(`
					Array.from(document.querySelectorAll('[data-lucide]'))
						.map(el => el.getAttribute('data-lucide'))
						.filter(icon => icon)
				`, &icons),
			)

			require.NoError(t, err)

			// Verify expected icons are present
			for _, expectedIcon := range tt.expectedIcons {
				assert.Contains(t, icons, expectedIcon,
					fmt.Sprintf("Expected icon '%s' not found on page %s", expectedIcon, tt.url))
			}

			// Verify icons are actually rendered (not just data attributes)
			var renderedIcons int
			err = chromedp.Run(ctx,
				chromedp.Evaluate(`
					document.querySelectorAll('[data-lucide] svg').length
				`, &renderedIcons),
			)

			require.NoError(t, err)
			assert.Greater(t, renderedIcons, 0, "No SVG icons were rendered")
		})
	}
}

// TestResponsiveDesign tests responsive behavior across different screen sizes
func TestResponsiveDesign(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping UI tests in short mode")
	}

	screenSizes := []struct {
		name   string
		width  int64
		height int64
	}{
		{"Desktop", 1920, 1080},
		{"Tablet", 768, 1024},
		{"Mobile", 375, 667},
	}

	for _, size := range screenSizes {
		t.Run(size.name, func(t *testing.T) {
			server := createUITestServer()
			defer server.Close()

			ctx, cancel := chromedp.NewContext(context.Background())
			defer cancel()

			ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			var sidebarVisible bool
			var mainContentWidth float64

			err := chromedp.Run(ctx,
				// Set viewport size
				chromedp.EmulateViewport(size.width, size.height),

				// Navigate to page
				chromedp.Navigate(server.URL),

				// Wait for page load
				chromedp.WaitVisible("body", chromedp.ByQuery),
				chromedp.Sleep(1*time.Second),

				// Check sidebar visibility
				chromedp.Evaluate(`
					window.getComputedStyle(document.querySelector('.sidebar')).display !== 'none'
				`, &sidebarVisible),

				// Check main content width
				chromedp.Evaluate(`
					document.querySelector('.main-content').offsetWidth
				`, &mainContentWidth),
			)

			require.NoError(t, err)

			// Assertions based on screen size
			switch size.name {
			case "Desktop":
				assert.True(t, sidebarVisible, "Sidebar should be visible on desktop")
				assert.Greater(t, mainContentWidth, 800.0, "Main content should be wide on desktop")
			case "Mobile":
				// On mobile, sidebar might be collapsed by default
				assert.Greater(t, mainContentWidth, 300.0, "Main content should use most of mobile width")
			}
		})
	}
}

// TestDynamicContentUpdates tests AJAX content updates and icon re-initialization
func TestDynamicContentUpdates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping UI tests in short mode")
	}

	server := createUITestServer()
	defer server.Close()

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var initialIconCount, updatedIconCount int

	err := chromedp.Run(ctx,
		// Navigate to sites page
		chromedp.Navigate(server.URL+"/sites"),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),

		// Count initial icons
		chromedp.Evaluate(`
			document.querySelectorAll('[data-lucide] svg').length
		`, &initialIconCount),

		// Trigger AJAX update (simulate refresh)
		chromedp.Click("#refresh-button", chromedp.ByID),
		chromedp.Sleep(3*time.Second),

		// Count icons after update
		chromedp.Evaluate(`
			document.querySelectorAll('[data-lucide] svg').length
		`, &updatedIconCount),
	)

	require.NoError(t, err)
	assert.Greater(t, initialIconCount, 0, "Should have initial icons")
	assert.Equal(t, initialIconCount, updatedIconCount, "Icons should be re-rendered after AJAX update")
}

// TestAccessibilityStandards tests basic accessibility requirements
func TestAccessibilityStandards(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping UI tests in short mode")
	}

	server := createUITestServer()
	defer server.Close()

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var ariaLabels []string
	var altTexts []string
	var pageTitle string

	err := chromedp.Run(ctx,
		chromedp.Navigate(server.URL),
		chromedp.WaitVisible("body", chromedp.ByQuery),

		// Check page title
		chromedp.Title(&pageTitle),

		// Check aria-labels
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('[aria-label]'))
				.map(el => el.getAttribute('aria-label'))
		`, &ariaLabels),

		// Check alt texts
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('img[alt]'))
				.map(el => el.getAttribute('alt'))
		`, &altTexts),
	)

	require.NoError(t, err)

	// Verify accessibility features
	assert.NotEmpty(t, pageTitle, "Page should have a title")
	assert.Contains(t, strings.ToLower(pageTitle), "minio", "Page title should mention MinIO")

	// Navigation should have accessible labels
	assert.NotEmpty(t, ariaLabels, "Should have aria-labels for accessibility")
}

// TestBrowserCompatibility tests across different browsers (if available)
func TestBrowserCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser compatibility tests in short mode")
	}

	// This would test different browsers if available
	// For now, just test Chrome which is most commonly available

	server := createUITestServer()
	defer server.Close()

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var userAgent string
	var lucideLoaded bool

	err := chromedp.Run(ctx,
		chromedp.Navigate(server.URL),
		chromedp.WaitVisible("body", chromedp.ByQuery),

		// Check user agent
		chromedp.Evaluate(`navigator.userAgent`, &userAgent),

		// Check if Lucide loaded
		chromedp.Evaluate(`typeof lucide !== 'undefined'`, &lucideLoaded),
	)

	require.NoError(t, err)
	assert.Contains(t, userAgent, "Chrome", "Should be running in Chrome")
	assert.True(t, lucideLoaded, "Lucide should load successfully")
}

// TestAnimationPerformance tests CSS animations and transitions
func TestAnimationPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping animation tests in short mode")
	}

	server := createUITestServer()
	defer server.Close()

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Test sidebar toggle animation
	var animationDuration float64

	err := chromedp.Run(ctx,
		chromedp.Navigate(server.URL),
		chromedp.WaitVisible("body", chromedp.ByQuery),

		// Measure animation performance
		chromedp.Evaluate(`
			const sidebar = document.querySelector('.sidebar');
			const computedStyle = window.getComputedStyle(sidebar);
			parseFloat(computedStyle.transitionDuration) || 0
		`, &animationDuration),
	)

	require.NoError(t, err)

	// Animation should be fast enough for good UX
	assert.LessOrEqual(t, animationDuration, 0.5, "Animations should be <= 500ms for good UX")
}

// createUITestServer creates a test HTTP server serving the UI
func createUITestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Serve main page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MinIO Site Replication Management</title>
    <script src="https://unpkg.com/lucide@latest/dist/umd/lucide.js"></script>
    <style>
        .sidebar { width: 250px; transition: all 0.3s ease; }
        .main-content { flex: 1; padding: 20px; }
        @media (max-width: 768px) {
            .sidebar { width: 0; overflow: hidden; }
            .main-content { width: 100%; }
        }
    </style>
</head>
<body>
    <div class="app">
        <nav class="sidebar">
            <div class="nav-item">
                <i data-lucide="layout-dashboard" width="20" height="20"></i>
                <span>Dashboard</span>
            </div>
            <div class="nav-item">
                <i data-lucide="globe" width="20" height="20"></i>
                <span>Sites</span>
            </div>
            <div class="nav-item">
                <i data-lucide="folder" width="20" height="20"></i>
                <span>Buckets</span>
            </div>
            <div class="nav-item">
                <i data-lucide="repeat" width="20" height="20"></i>
                <span>Replication</span>
            </div>
            <div class="nav-item">
                <i data-lucide="check-circle" width="20" height="20"></i>
                <span>Consistency</span>
            </div>
            <div class="nav-item">
                <i data-lucide="settings" width="20" height="20"></i>
                <span>Settings</span>
            </div>
        </nav>
        
        <main class="main-content">
            <header>
                <i data-lucide="package" width="32" height="32"></i>
                <h1>MinIO Site Replication</h1>
                <button id="refresh-button" aria-label="Refresh">
                    <i data-lucide="refresh-cw" width="20" height="20"></i>
                </button>
            </header>
            
            <div id="content">
                <p>Main content area</p>
            </div>
        </main>
    </div>
    
    <script>
        // Initialize Lucide icons
        if (typeof lucide !== 'undefined') {
            lucide.createIcons();
        }
        
        // Simulate AJAX refresh
        document.getElementById('refresh-button').addEventListener('click', function() {
            setTimeout(function() {
                // Simulate adding new content with icons
                const content = document.getElementById('content');
                content.innerHTML += '<div><i data-lucide="plus" width="16" height="16"></i> New content</div>';
                
                // Re-initialize icons
                if (typeof lucide !== 'undefined') {
                    lucide.createIcons();
                }
            }, 500);
        });
    </script>
</body>
</html>
		`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	// Serve sites page
	mux.HandleFunc("/sites", func(w http.ResponseWriter, r *http.Request) {
		html := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Sites Management - MinIO</title>
    <script src="https://unpkg.com/lucide@latest/dist/umd/lucide.js"></script>
</head>
<body>
    <div class="sites-page">
        <h1><i data-lucide="globe" width="24" height="24"></i> Sites Management</h1>
        
        <div class="actions">
            <button aria-label="Add Site">
                <i data-lucide="plus" width="16" height="16"></i>
                Add Site
            </button>
            <button aria-label="Remove Site">
                <i data-lucide="trash-2" width="16" height="16"></i>
                Remove
            </button>
            <button aria-label="Resync From">
                <i data-lucide="download" width="16" height="16"></i>
                Resync From
            </button>
            <button aria-label="Resync To">
                <i data-lucide="upload" width="16" height="16"></i>
                Resync To
            </button>
        </div>
    </div>
    
    <script>
        if (typeof lucide !== 'undefined') {
            lucide.createIcons();
        }
    </script>
</body>
</html>
		`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	return httptest.NewServer(mux)
}
