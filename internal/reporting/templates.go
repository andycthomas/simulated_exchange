package reporting

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"sync"
	"time"
)

// StandardTemplateEngine implements TemplateEngine interface
type StandardTemplateEngine struct {
	templates map[string]*template.Template
	mutex     sync.RWMutex
	funcMap   template.FuncMap
}

// NewStandardTemplateEngine creates a new template engine
func NewStandardTemplateEngine() *StandardTemplateEngine {
	engine := &StandardTemplateEngine{
		templates: make(map[string]*template.Template),
		funcMap:   createDefaultFuncMap(),
	}

	// Register default templates
	engine.registerDefaultTemplates()

	return engine
}

// RenderReport generates formatted report from template
func (ste *StandardTemplateEngine) RenderReport(templateName string, data interface{}) ([]byte, error) {
	ste.mutex.RLock()
	tmpl, exists := ste.templates[templateName]
	ste.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("template '%s' not found", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// RegisterTemplate adds new template to engine
func (ste *StandardTemplateEngine) RegisterTemplate(name string, templateContent string) error {
	tmpl, err := template.New(name).Funcs(ste.funcMap).Parse(templateContent)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	ste.mutex.Lock()
	ste.templates[name] = tmpl
	ste.mutex.Unlock()

	return nil
}

// GetAvailableTemplates returns list of available templates
func (ste *StandardTemplateEngine) GetAvailableTemplates() []string {
	ste.mutex.RLock()
	defer ste.mutex.RUnlock()

	templates := make([]string, 0, len(ste.templates))
	for name := range ste.templates {
		templates = append(templates, name)
	}

	return templates
}

// ValidateTemplate checks template syntax
func (ste *StandardTemplateEngine) ValidateTemplate(templateContent string) error {
	_, err := template.New("validation").Funcs(ste.funcMap).Parse(templateContent)
	return err
}

// registerDefaultTemplates registers built-in templates
func (ste *StandardTemplateEngine) registerDefaultTemplates() {
	templates := map[string]string{
		"executive_report": getExecutiveReportTemplate(),
		"roi_analysis":     getROIAnalysisTemplate(),
		"risk_assessment":  getRiskAssessmentTemplate(),
		"performance_summary": getPerformanceSummaryTemplate(),
		"cost_optimization": getCostOptimizationTemplate(),
	}

	for name, content := range templates {
		if err := ste.RegisterTemplate(name, content); err != nil {
			// Log error but continue with other templates
			fmt.Printf("Warning: Failed to register template '%s': %v\n", name, err)
		}
	}
}

// createDefaultFuncMap creates template functions
func createDefaultFuncMap() template.FuncMap {
	return template.FuncMap{
		"formatCurrency": func(amount float64, currency string) string {
			return fmt.Sprintf("%s %.2f", currency, amount)
		},
		"formatPercent": func(value float64) string {
			return fmt.Sprintf("%.1f%%", value)
		},
		"formatNumber": func(value float64) string {
			return fmt.Sprintf("%.2f", value)
		},
		"formatDate": func(date time.Time) string {
			return date.Format("January 2, 2006")
		},
		"formatShortDate": func(date time.Time) string {
			return date.Format("2006-01-02")
		},
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": strings.Title,
		"add": func(a, b float64) float64 {
			return a + b
		},
		"subtract": func(a, b float64) float64 {
			return a - b
		},
		"multiply": func(a, b float64) float64 {
			return a * b
		},
		"divide": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"round": func(value float64, decimals int) float64 {
			factor := 1.0
			for i := 0; i < decimals; i++ {
				factor *= 10
			}
			return float64(int(value*factor+0.5)) / factor
		},
		"gradeColor": func(grade Grade) string {
			switch grade {
			case GradeExcellent:
				return "#22c55e" // Green
			case GradeGood:
				return "#84cc16" // Light green
			case GradeSatisfactory:
				return "#eab308" // Yellow
			case GradeNeedsImprovement:
				return "#f97316" // Orange
			case GradePoor:
				return "#ef4444" // Red
			case GradeCritical:
				return "#dc2626" // Dark red
			default:
				return "#6b7280" // Gray
			}
		},
		"trendIcon": func(trend TrendDirection) string {
			switch trend {
			case TrendUp:
				return "↗️"
			case TrendDown:
				return "↘️"
			case TrendStable:
				return "→"
			case TrendVolatile:
				return "↕️"
			default:
				return "?"
			}
		},
		"riskColor": func(level string) string {
			switch strings.ToLower(level) {
			case "low":
				return "#22c55e"
			case "medium":
				return "#eab308"
			case "high":
				return "#ef4444"
			case "critical":
				return "#dc2626"
			default:
				return "#6b7280"
			}
		},
		"progressBar": func(current, max float64, width int) string {
			if max == 0 {
				return strings.Repeat("□", width)
			}
			progress := int((current / max) * float64(width))
			if progress > width {
				progress = width
			}
			return strings.Repeat("■", progress) + strings.Repeat("□", width-progress)
		},
		"slice": func(items interface{}, start, end int) interface{} {
			// This would need type assertion in real implementation
			return items
		},
	}
}

// Template definitions

func getExecutiveReportTemplate() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Executive Report - {{.Metadata.Title}}</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 20px;
            background-color: #f8fafc;
            color: #1f2937;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 8px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
            overflow: hidden;
        }

        .header {
            background: linear-gradient(135deg, #1f2937 0%, #374151 100%);
            color: white;
            padding: 40px;
            text-align: center;
        }

        .header h1 {
            margin: 0;
            font-size: 2.5em;
            font-weight: 300;
        }

        .header .subtitle {
            margin-top: 10px;
            opacity: 0.8;
            font-size: 1.1em;
        }

        .section {
            padding: 30px 40px;
            border-bottom: 1px solid #e5e7eb;
        }

        .section:last-child {
            border-bottom: none;
        }

        .section h2 {
            color: #1f2937;
            margin-bottom: 20px;
            font-size: 1.8em;
            border-bottom: 2px solid #3b82f6;
            padding-bottom: 10px;
        }

        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin: 20px 0;
        }

        .metric-card {
            background: #f8fafc;
            border: 1px solid #e5e7eb;
            border-radius: 8px;
            padding: 20px;
            text-align: center;
        }

        .metric-value {
            font-size: 2em;
            font-weight: bold;
            margin-bottom: 5px;
        }

        .metric-label {
            color: #6b7280;
            font-size: 0.9em;
        }

        .score-card {
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 20px 0;
        }

        .score-circle {
            width: 120px;
            height: 120px;
            border-radius: 50%;
            border: 8px solid {{gradeColor .ExecutiveSummary.OverallScore.Grade}};
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            background: white;
            margin-right: 30px;
        }

        .score-number {
            font-size: 2em;
            font-weight: bold;
            color: {{gradeColor .ExecutiveSummary.OverallScore.Grade}};
        }

        .score-grade {
            font-size: 1.2em;
            color: {{gradeColor .ExecutiveSummary.OverallScore.Grade}};
        }

        .findings-list {
            list-style: none;
            padding: 0;
        }

        .findings-list li {
            background: #f0f9ff;
            border-left: 4px solid #3b82f6;
            margin: 10px 0;
            padding: 15px;
            border-radius: 0 4px 4px 0;
        }

        .recommendations-list {
            list-style: none;
            padding: 0;
        }

        .recommendations-list li {
            background: #f0fdf4;
            border-left: 4px solid #22c55e;
            margin: 10px 0;
            padding: 15px;
            border-radius: 0 4px 4px 0;
        }

        .trend-indicator {
            display: inline-block;
            margin-left: 10px;
            font-size: 1.2em;
        }

        .footer {
            background: #f9fafb;
            padding: 20px 40px;
            text-align: center;
            color: #6b7280;
            font-size: 0.9em;
        }

        @media print {
            body { background: white; }
            .container { box-shadow: none; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Metadata.Title}}</h1>
            <div class="subtitle">
                {{.Metadata.BusinessUnit}} | {{formatDate .Metadata.GeneratedAt}}
            </div>
        </div>

        <div class="section">
            <h2>Executive Summary</h2>
            <div class="score-card">
                <div class="score-circle">
                    <div class="score-number">{{round .ExecutiveSummary.OverallScore.Score 1}}</div>
                    <div class="score-grade">{{.ExecutiveSummary.OverallScore.Grade}}</div>
                </div>
                <div>
                    <h3>Overall Performance Assessment</h3>
                    <p>{{.ExecutiveSummary.BusinessImpact}}</p>
                </div>
            </div>

            <h3>Key Findings</h3>
            <ul class="findings-list">
                {{range .ExecutiveSummary.KeyFindings}}
                <li>{{.Finding}} <strong>(Impact: {{.Impact}})</strong></li>
                {{end}}
            </ul>

            <h3>Top Recommendations</h3>
            <ul class="recommendations-list">
                {{range .ExecutiveSummary.TopRecommendations}}
                <li>{{.}}</li>
                {{end}}
            </ul>
        </div>

        <div class="section">
            <h2>Key Performance Indicators</h2>
            <div class="metrics-grid">
                <div class="metric-card">
                    <div class="metric-value" style="color: {{gradeColor .PerformanceSection.OverallPerformance.PerformanceGrade}};">
                        {{formatPercent .PerformanceSection.KPIMetrics.CustomerSatisfaction}}
                    </div>
                    <div class="metric-label">Customer Satisfaction</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value" style="color: {{gradeColor .PerformanceSection.OverallPerformance.PerformanceGrade}};">
                        {{formatPercent .PerformanceSection.KPIMetrics.EmployeeEngagement}}
                    </div>
                    <div class="metric-label">Employee Engagement</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value" style="color: {{gradeColor .PerformanceSection.OverallPerformance.PerformanceGrade}};">
                        {{formatPercent .PerformanceSection.KPIMetrics.OperationalEfficiency}}
                    </div>
                    <div class="metric-label">Operational Efficiency</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value" style="color: {{gradeColor .PerformanceSection.OverallPerformance.PerformanceGrade}};">
                        {{formatPercent .PerformanceSection.KPIMetrics.MarketShare}}
                    </div>
                    <div class="metric-label">Market Share</div>
                </div>
            </div>
        </div>

        <div class="section">
            <h2>ROI Analysis</h2>
            <div class="metrics-grid">
                <div class="metric-card">
                    <div class="metric-value">{{formatCurrency .ROIAnalysis.ROICalculation.InitialInvestment "USD"}}</div>
                    <div class="metric-label">Initial Investment</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value">{{formatPercent .ROIAnalysis.ROICalculation.ThreeYearROI}}</div>
                    <div class="metric-label">3-Year ROI</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value">{{round .ROIAnalysis.ROICalculation.PaybackPeriod 1}} months</div>
                    <div class="metric-label">Payback Period</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value">{{formatPercent .ROIAnalysis.IRRAnalysis.InternalRateOfReturn}}</div>
                    <div class="metric-label">Internal Rate of Return</div>
                </div>
            </div>
        </div>

        <div class="section">
            <h2>Risk Assessment</h2>
            <p><strong>Overall Risk Level:</strong>
                <span style="color: {{riskColor .RiskMitigation.RiskAnalysis.RiskLevel}};">
                    {{upper .RiskMitigation.RiskAnalysis.RiskLevel}}
                </span>
            </p>
            <p><strong>Risk Score:</strong> {{round .RiskMitigation.RiskAnalysis.OverallRiskScore 1}}/100</p>

            <h3>Top Risks</h3>
            <ul>
                {{range .RiskMitigation.RiskAnalysis.TopRisks}}
                <li><strong>{{.Name}}:</strong> {{.Description}} (Probability: {{formatPercent .Probability}}, Impact: {{.Impact}})</li>
                {{end}}
            </ul>
        </div>

        <div class="footer">
            Generated by {{.Metadata.GeneratedBy}} on {{formatShortDate .Metadata.GeneratedAt}} |
            Version {{.Metadata.Version}} | Classification: {{.Metadata.Classification}}
        </div>
    </div>
</body>
</html>`
}

func getROIAnalysisTemplate() string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>ROI Analysis Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f4f4f4; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; }
        .metric { display: inline-block; margin: 10px; padding: 15px; background: #e9e9e9; border-radius: 5px; }
        .roi-table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        .roi-table th, .roi-table td { border: 1px solid #ddd; padding: 8px; text-align: right; }
        .roi-table th { background: #f2f2f2; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Return on Investment Analysis</h1>
        <p>Investment ID: {{.InvestmentID}}</p>
    </div>

    <div class="section">
        <h2>Investment Summary</h2>
        <div class="metric">
            <strong>Initial Investment:</strong> {{formatCurrency .ROICalculation.InitialInvestment "USD"}}
        </div>
        <div class="metric">
            <strong>Payback Period:</strong> {{round .PaybackAnalysis.SimplePaybackMonths 1}} months
        </div>
        <div class="metric">
            <strong>Total ROI:</strong> {{formatPercent .ROICalculation.TotalROI}}
        </div>
        <div class="metric">
            <strong>IRR:</strong> {{formatPercent .IRRAnalysis.InternalRateOfReturn}}
        </div>
    </div>

    <div class="section">
        <h2>Annual Benefits Projection</h2>
        <table class="roi-table">
            <thead>
                <tr>
                    <th>Year</th>
                    <th>Annual Benefits</th>
                    <th>Cumulative Benefits</th>
                    <th>ROI to Date</th>
                </tr>
            </thead>
            <tbody>
                {{range $index, $benefit := .ROICalculation.AnnualBenefits}}
                <tr>
                    <td>{{add $index 1}}</td>
                    <td>{{formatCurrency $benefit "USD"}}</td>
                    <td>{{formatCurrency (index $.CashFlowProjection.CumulativeCashFlows $index) "USD"}}</td>
                    <td>{{formatPercent (divide (subtract (index $.CashFlowProjection.CumulativeCashFlows $index) $.ROICalculation.InitialInvestment) $.ROICalculation.InitialInvestment | multiply 100)}}</td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>

    <div class="section">
        <h2>Risk Assessment</h2>
        <p><strong>Investment Risk Level:</strong> {{.RiskLevel}}</p>
        <p><strong>Confidence Level:</strong> {{formatPercent .IRRAnalysis.ConfidenceLevel}}</p>
        <p><strong>Recommendation:</strong> {{.IRRAnalysis.Recommendation}}</p>
    </div>
</body>
</html>`
}

func getRiskAssessmentTemplate() string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>Risk Assessment Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .risk-high { color: #dc2626; }
        .risk-medium { color: #f59e0b; }
        .risk-low { color: #10b981; }
        .risk-matrix { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; margin: 20px 0; }
        .risk-cell { padding: 15px; text-align: center; border-radius: 5px; }
    </style>
</head>
<body>
    <h1>Risk Assessment Report</h1>

    <div class="section">
        <h2>Overall Risk Profile</h2>
        <p><strong>Risk Score:</strong> {{round .OverallRiskScore 1}}/100</p>
        <p><strong>Risk Level:</strong> <span class="risk-{{lower .RiskLevel}}">{{upper .RiskLevel}}</span></p>
    </div>

    <div class="section">
        <h2>Top Risks</h2>
        <ol>
            {{range .TopRisks}}
            <li>
                <strong>{{.Name}}</strong> ({{.Category}})
                <br>{{.Description}}
                <br><em>Probability: {{formatPercent .Probability}} | Impact: {{.Impact}}</em>
            </li>
            {{end}}
        </ol>
    </div>

    <div class="section">
        <h2>Risk Mitigation Strategies</h2>
        {{range .MitigationStrategies}}
        <div style="margin: 15px 0; padding: 15px; background: #f9f9f9; border-radius: 5px;">
            <h4>{{.Title}}</h4>
            <p>{{.Description}}</p>
            <p><strong>Timeline:</strong> {{.Timeline}} | <strong>Cost:</strong> {{formatCurrency .EstimatedCost "USD"}}</p>
        </div>
        {{end}}
    </div>
</body>
</html>`
}

func getPerformanceSummaryTemplate() string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>Performance Summary</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .kpi-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; }
        .kpi-card { padding: 20px; background: #f8f9fa; border-radius: 8px; text-align: center; }
        .kpi-value { font-size: 2em; font-weight: bold; margin-bottom: 5px; }
        .trend { font-size: 1.5em; }
    </style>
</head>
<body>
    <h1>Performance Summary</h1>

    <div class="kpi-grid">
        <div class="kpi-card">
            <div class="kpi-value">{{formatPercent .KPIMetrics.CustomerSatisfaction}}</div>
            <div>Customer Satisfaction</div>
            <div class="trend">{{trendIcon .TrendAnalysis.CustomerTrend}}</div>
        </div>
        <div class="kpi-card">
            <div class="kpi-value">{{formatPercent .KPIMetrics.EmployeeEngagement}}</div>
            <div>Employee Engagement</div>
            <div class="trend">{{trendIcon .TrendAnalysis.EmployeeTrend}}</div>
        </div>
        <div class="kpi-card">
            <div class="kpi-value">{{formatPercent .KPIMetrics.OperationalEfficiency}}</div>
            <div>Operational Efficiency</div>
            <div class="trend">{{trendIcon .TrendAnalysis.OperationalTrend}}</div>
        </div>
        <div class="kpi-card">
            <div class="kpi-value">{{formatPercent .KPIMetrics.MarketShare}}</div>
            <div>Market Share</div>
            <div class="trend">{{trendIcon .TrendAnalysis.MarketTrend}}</div>
        </div>
    </div>

    <div class="section">
        <h2>Performance Analysis</h2>
        <p><strong>Overall Grade:</strong> {{.OverallPerformance.PerformanceGrade}}</p>
        <p><strong>Performance Score:</strong> {{round .OverallPerformance.OverallScore 1}}/100</p>
        <p><strong>Trend Direction:</strong> {{.OverallPerformance.TrendDirection}} {{trendIcon .OverallPerformance.TrendDirection}}</p>
    </div>
</body>
</html>`
}

func getCostOptimizationTemplate() string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>Cost Optimization Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .savings-card { background: #ecfdf5; border: 1px solid #10b981; border-radius: 8px; padding: 20px; margin: 15px 0; }
        .cost-breakdown { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 15px; }
        .cost-category { padding: 15px; background: #f3f4f6; border-radius: 5px; }
    </style>
</head>
<body>
    <h1>Cost Optimization Analysis</h1>

    <div class="section">
        <h2>Current Cost Breakdown</h2>
        <div class="cost-breakdown">
            {{range $category, $amount := .CostBreakdown.Categories}}
            <div class="cost-category">
                <h4>{{title $category}}</h4>
                <div style="font-size: 1.5em; font-weight: bold;">{{formatCurrency $amount "USD"}}</div>
                <div>{{formatPercent (divide $amount $.CostBreakdown.TotalCosts | multiply 100)}} of total</div>
            </div>
            {{end}}
        </div>
    </div>

    <div class="section">
        <h2>Savings Opportunities</h2>
        {{range .SavingsOpportunities}}
        <div class="savings-card">
            <h3>{{.Title}}</h3>
            <p>{{.Description}}</p>
            <p><strong>Potential Savings:</strong> {{formatCurrency .PotentialSavings "USD"}} annually</p>
            <p><strong>Implementation Effort:</strong> {{.ImplementationEffort}}</p>
            <p><strong>Timeline:</strong> {{.Timeline}}</p>
        </div>
        {{end}}
    </div>

    <div class="section">
        <h2>Efficiency Metrics</h2>
        <p><strong>Overall Efficiency Score:</strong> {{round .EfficiencyMetrics.OverallEfficiency 1}}/100</p>
        <p><strong>Cost per Unit:</strong> {{formatCurrency .EfficiencyMetrics.CostPerUnit "USD"}}</p>
        <p><strong>Productivity Index:</strong> {{round .EfficiencyMetrics.ProductivityIndex 2}}</p>
    </div>
</body>
</html>`
}

// Supporting template data structures would be defined here
// These represent the data structures expected by the templates