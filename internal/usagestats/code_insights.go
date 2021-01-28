package usagestats

import "context"

func GetCodeInsightsUsageStatistics(ctx context.Context) {
	// Viewing metrics

	const viewingMetricsQuery = `
	SELECT
		COUNT(*) FILTER (WHERE name = 'ViewInsights')                       AS insights_page_views,
		COUNT(distinct user_id) FILTER (WHERE name = 'ViewInsights')        AS insights_unique_page_views,
		COUNT(*) FILTER
			(WHERE name = 'InsightHover'
				AND argument ->> 'insightType'::text = 'searchInsights')    AS search_insights_hovers,
		COUNT(*) FILTER
			(WHERE name = 'InsightHover'
				AND argument ->> 'insightType'::text = 'codeStatsInsights') AS code_stats_insights_hovers,
		COUNT(*) FILTER (WHERE name = 'InsightUICustomization')             AS insights_ui_customizations,
		COUNT(*) FILTER (WHERE name = 'InsightDataPointClick')              AS insights_data_point_clicks
	FROM event_logs
	WHERE name in ('ViewInsights', 'InsightHover', 'InsightUICustomization', 'InsightDataPointClick');
	`

	// 'ViewInsights' by day,
	// unique users in views DAU/WAU/MAU
	// 'InsightHover' by type and day
	// 'InsightUICustomization' graph resize + moves
	// 'InsightDataPointClick' graph data point click

	// Creation metrics

	// InsightEdit
	// InsightAddition
	// InsightRemoval
	// users who have added or edited a code insight
	// users who have added or edited a code insight for the first time this week

	// InsightConfigureClick
	// InsightAddMoreClick
}
