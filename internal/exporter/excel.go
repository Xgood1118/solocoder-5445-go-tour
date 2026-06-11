package exporter

import (
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"

	"tourtool/internal/product"
	"tourtool/internal/quote"
	"tourtool/pkg/i18n"
)

type ExportVersion string

const (
	VersionSales    ExportVersion = "sales"
	VersionCustomer ExportVersion = "customer"
)

type ExcelExporter struct {
	lang    i18n.Lang
	version ExportVersion
}

func NewExcelExporter(lang i18n.Lang, version ExportVersion) *ExcelExporter {
	return &ExcelExporter{lang: lang, version: version}
}

func (e *ExcelExporter) Export(p *product.Product, q *quote.QuoteResult, outputPath string) error {
	i18n.SetLang(e.lang)

	f := excelize.NewFile()
	defer f.Close()

	sheetName := i18n.T("quotation_sheet")
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}
	f.DeleteSheet("Sheet1")
	f.SetActiveSheet(index)

	e.setColumnWidths(f, sheetName)

	row := 1

	row = e.writeHeader(f, sheetName, row)

	row += 1
	row = e.writeProductInfo(f, sheetName, row, p, q)

	row += 1
	row = e.writePriceInfo(f, sheetName, row, q)

	row += 1
	row = e.writeItinerary(f, sheetName, row, p)

	row += 1
	row = e.writeInclusionExclusion(f, sheetName, row)

	if e.version == VersionCustomer {
		e.addWatermark(f, sheetName)
		e.addValidUntil(f, sheetName)
	}

	e.applyStyles(f, sheetName)

	return f.SaveAs(outputPath)
}

func (e *ExcelExporter) setColumnWidths(f *excelize.File, sheet string) {
	widths := map[string]float64{
		"A": 15,
		"B": 25,
		"C": 20,
		"D": 15,
		"E": 15,
		"F": 20,
		"G": 20,
	}
	for col, w := range widths {
		f.SetColWidth(sheet, col, col, w)
	}
}

func (e *ExcelExporter) writeHeader(f *excelize.File, sheet string, row int) int {
	title := i18n.T("quotation_sheet")
	versionLabel := i18n.T("sales_version")
	if e.version == VersionCustomer {
		versionLabel = i18n.T("customer_version")
	}

	f.MergeCell(sheet, "A1", "G1")
	f.SetCellValue(sheet, "A1", fmt.Sprintf("%s - %s", title, versionLabel))

	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 18,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	f.SetCellStyle(sheet, "A1", "G1", style)
	f.SetRowHeight(sheet, 1, 40)

	f.SetCellValue(sheet, "G2", i18n.T("company_logo"))
	f.SetCellValue(sheet, "F3", i18n.T("seal"))
	f.SetCellValue(sheet, "G3", i18n.T("qr_code"))

	return 4
}

func (e *ExcelExporter) writeProductInfo(f *excelize.File, sheet string, row int, p *product.Product, q *quote.QuoteResult) int {
	startRow := row

	headers := []string{i18n.T("product_name"), i18n.T("destination"), i18n.T("days"), i18n.T("nights"), i18n.T("target_audience"), i18n.T("tier")}
	for i, h := range headers {
		cell := fmt.Sprintf("%c%d", 'A'+i, row)
		f.SetCellValue(sheet, cell, h)
	}
	row++

	audiences := ""
	for i, a := range p.TargetAudience {
		if i > 0 {
			audiences += ", "
		}
		audiences += i18n.T(string(a))
	}

	tierLabel := i18n.T(p.Tier)

	values := []interface{}{p.Name, p.Destination, p.Days, p.Nights, audiences, tierLabel}
	for i, v := range values {
		cell := fmt.Sprintf("%c%d", 'A'+i, row)
		f.SetCellValue(sheet, cell, v)
	}
	row++

	e.applyHeaderStyle(f, sheet, startRow, startRow)

	return row
}

func (e *ExcelExporter) writePriceInfo(f *excelize.File, sheet string, row int, q *quote.QuoteResult) int {
	f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), i18n.T("price_adjustment"))
	row++

	priceItems := []struct {
		label string
		value float64
		show  bool
	}{
		{i18n.T("single_price"), q.SinglePrice, true},
		{i18n.T("double_price"), q.DoublePrice, true},
		{i18n.T("single_room_diff"), q.SingleRoomDiff, true},
		{i18n.T("child_price_no_bed"), q.ChildNoBedPrice, q.HasChildPrices},
		{i18n.T("child_price_with_bed"), q.ChildWithBedPrice, q.HasChildPrices},
		{i18n.T("senior_price"), q.SeniorPrice, q.HasSeniorPrices},
		{i18n.T("companion_price"), q.CompanionPrice, q.HasSeniorPrices},
	}

	for _, item := range priceItems {
		if !item.show {
			continue
		}
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), item.label)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("¥%.2f", item.value))
		row++
	}

	return row
}

func (e *ExcelExporter) writeItinerary(f *excelize.File, sheet string, row int, p *product.Product) int {
	f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), i18n.T("itinerary"))
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E8F4FD"}, Pattern: 1},
	})
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), headerStyle)
	row++

	dayHeaders := []string{
		i18n.T("day") + i18n.T("day_suffix"),
		i18n.T("hotel"),
		i18n.T("sight"),
		i18n.T("breakfast"),
		i18n.T("lunch"),
		i18n.T("dinner"),
		i18n.T("transport"),
	}
	for i, h := range dayHeaders {
		cell := fmt.Sprintf("%c%d", 'A'+i, row)
		f.SetCellValue(sheet, cell, h)
	}
	e.applyHeaderStyle(f, sheet, row, row)
	row++

	for _, day := range p.Itinerary {
		dayLabel := ""
		switch day.DayType {
		case product.DayDeparture:
			dayLabel = i18n.T("departure_day")
		case product.DaySight:
			dayLabel = i18n.T("tour_day")
		case product.DayReturn:
			dayLabel = i18n.T("return_day")
		}

		sights := ""
		for i, s := range day.SightIDs {
			if i > 0 {
				sights += ", "
			}
			sights += s
		}

		bf := day.Breakfast
		if bf == "included" {
			bf = "✓"
		}
		lc := day.Lunch
		if lc == "included" {
			lc = "✓"
		}
		dn := day.Dinner
		if dn == "included" {
			dn = "✓"
		}

		transport := "—"
		if day.BusIncluded {
			transport = i18n.T("tour_bus")
		}

		values := []interface{}{
			fmt.Sprintf("Day %d %s", day.DayNumber, dayLabel),
			string(day.HotelTier),
			sights,
			bf,
			lc,
			dn,
			transport,
		}
		for i, v := range values {
			cell := fmt.Sprintf("%c%d", 'A'+i, row)
			f.SetCellValue(sheet, cell, v)
		}
		row++
	}

	return row
}

func (e *ExcelExporter) writeInclusionExclusion(f *excelize.File, sheet string, row int) int {
	startRow := row

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#D5E8D4"}, Pattern: 1},
	})
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), headerStyle)
	exclStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#F8CECC"}, Pattern: 1},
	})
	f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("G%d", row), exclStyle)
	row++

	includes := []string{
		i18n.T("flight_train"),
		i18n.T("accommodation"),
		i18n.T("meals"),
		i18n.T("first_gate_ticket"),
		i18n.T("tour_bus"),
		i18n.T("guide_service"),
		i18n.T("insurance"),
	}
	excludes := []string{
		i18n.T("self_paid_items"),
		i18n.T("single_room_supplement"),
		i18n.T("excess_baggage"),
		i18n.T("personal_expenses"),
	}

	maxLen := len(includes)
	if len(excludes) > maxLen {
		maxLen = len(excludes)
	}

	for i := 0; i < maxLen; i++ {
		if i < len(includes) {
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("• %s", includes[i]))
		}
		if i < len(excludes) {
			f.SetCellValue(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("• %s", excludes[i]))
		}
		row++
	}

	return row
}

func (e *ExcelExporter) addWatermark(f *excelize.File, sheet string) {
	f.MergeCell(sheet, "A10", "G20")
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:  48,
			Color: "#CCCCCC",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	f.SetCellValue(sheet, "A10", "CONFIDENTIAL")
	f.SetCellStyle(sheet, "A10", "G20", style)
}

func (e *ExcelExporter) addValidUntil(f *excelize.File, sheet string) {
	validUntil := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	f.SetCellValue(sheet, "F30", i18n.T("valid_until"))
	f.SetCellValue(sheet, "G30", validUntil)
}

func (e *ExcelExporter) applyHeaderStyle(f *excelize.File, sheet string, startRow, endRow int) {
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#D9E2F3"}, Pattern: 1},
	})
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", startRow), fmt.Sprintf("G%d", endRow), style)
}

func (e *ExcelExporter) applyStyles(f *excelize.File, sheet string) {
	borderStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	f.SetCellStyle(sheet, "A1", "G50", borderStyle)
}
