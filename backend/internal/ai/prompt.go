package ai

import (
	"fmt"
	"strings"

	"meituan-aicoding/backend/internal/model"
)

func ReviewAnalysisInstruction() string {
	return strings.TrimSpace(`你是一名美食团购评价分析助手，请严格根据用户输入的评论内容完成分析。

你必须严格按以下标签顺序输出，不能输出 JSON，不能输出 markdown，不能输出额外解释，不能更改标签名称：
[SUMMARY]
一句中文总结
[/SUMMARY]
[POSITIVE_KEYWORDS]
- 关键词1
- 关键词2
- 关键词3
- 关键词4
- 关键词5
[/POSITIVE_KEYWORDS]
[NEGATIVE_KEYWORDS]
- 关键词1
- 关键词2
- 关键词3
- 关键词4
- 关键词5
[/NEGATIVE_KEYWORDS]
[SENTIMENT_SCORE]
0到100的整数
[/SENTIMENT_SCORE]
[SUGGESTIONS]
- 建议1
- 建议2
- 建议3
[/SUGGESTIONS]

补充要求：
1. 所有内容必须使用中文。
2. 关键词最多 5 个，可以少于 5 个。
3. 建议输出 2 到 3 条。
4. summary 必须是一句简洁中文总结。
5. sentiment_score 只能输出整数。
6. 只围绕商家经营视角分析，不要编造输入中没有的信息。`)
}

func BuildReviewAnalysisPrompt(reviews []model.Review, tab string) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("当前分析范围：%s。\n", tab))
	builder.WriteString("以下是用户评价列表，请聚合分析：\n")
	for i, review := range reviews {
		builder.WriteString(fmt.Sprintf("%d. 用户=%s；评分=%d；评论=%s\n", i+1, review.Username, review.Score, strings.TrimSpace(review.Content)))
	}
	return builder.String()
}

func ReviewReplyInstruction() string {
	return strings.TrimSpace(`你是一名美食商家运营助手，请针对单条差评生成商家回复话术。

要求：
1. 输出纯文本，不要 markdown。
2. 语气真诚、克制、专业。
3. 要包含致歉、问题回应、改进承诺。
4. 不要承诺退款或赠品。
5. 长度控制在 60 到 120 字。
6. 必须使用中文。`)
}

func BuildReplyPrompt(review *model.Review) string {
	return fmt.Sprintf("请为这条评价生成商家回复。用户=%s；评分=%d；评论=%s", review.Username, review.Score, strings.TrimSpace(review.Content))
}
