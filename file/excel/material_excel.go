package excelutil

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go_file/common"
	"go_file/utils/sliceutil"
	"go_file/utils/timeutil"

	"github.com/extrame/xls"
	"github.com/gogs/chardet"
	"github.com/tealeg/xlsx"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type ExcelFile struct {
	FileName string        // 文件名
	Sheets   []*ExcelSheet // 表单信息
	TotalRow int
}

type ExcelSheet struct {
	SheetName string     // 表单名称
	Header    []string   // 表头
	Rows      [][]string // 数据
}

// ReadExcelFile 读取Excel文件, 提取指定表头数据
// dstTitleMap 待提取的表头和要转为的目标表头映射
func ReadExcelFile(fileName string, checkTitles []string, dstTitleMap map[string]string) (
	*ExcelFile, error) {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case common.FileTypeXlsx:
		return ProcessXLSXFile(fileName, checkTitles, dstTitleMap)
	case common.FileTypeXls:
		return ProcessXLSFile(fileName, checkTitles, dstTitleMap)
	case common.FileTypeCsv:
		return ProcessCSVFile(fileName, checkTitles, dstTitleMap)
	}
	return nil, nil
}

// ProcessXLSXFile 处理 .xlsx 文件, 提取指定表头数据
// dstTitleMap 待提取的表头
// delNullCell 是否删除空单元格的行
func ProcessXLSXFile(fileName string, checkTitles []string, dstTitleMap map[string]string) (
	*ExcelFile, error) {
	xlFile, err := xlsx.OpenFile(fileName)
	if err != nil {
		return nil, err
	}
	retSheets := make([]*ExcelSheet, 0)
	totalRow := 0
	for _, sheet := range xlFile.Sheets {
		var excelSheet ExcelSheet
		excelSheet.SheetName = sheet.Name
		header, err := extractHeader(sheet.Rows)
		if err != nil {
			return nil, err
		}
		// 标记每一列是否有非空数据,并检测表头是否符合要求
		columnIsEmpty, err := checkExcelTitle(header, checkTitles)
		if err != nil {
			return nil, err
		}
		// 遍历所有行，检测列是否有非空数据
		for _, row := range sheet.Rows[1:] {
			for i, cell := range row.Cells {
				if cell.String() != "" {
					columnIsEmpty[i] = false
				}
			}
		}
		// 根据数据是否为空更新 dstTitleMap
		for i, empty := range columnIsEmpty {
			if empty {
				delete(dstTitleMap, header[i])
			}
		}
		for _, title := range header {
			if dst, ok := dstTitleMap[title]; ok {
				excelSheet.Header = append(excelSheet.Header, dst)
			}
		}
		// 处理每一行数据,提取需要的列数据
		for _, row := range sheet.Rows[1:] { // 跳过表头
			mappedData := mapRowData(header, row.Cells, dstTitleMap)
			if len(mappedData) > 0 {
				rowData := make([]string, 0)
				for _, title := range excelSheet.Header {
					if value, ok := mappedData[title]; ok {
						rowData = append(rowData, value)
					}
				}
				totalRow++
				excelSheet.Rows = append(excelSheet.Rows, rowData)
			}
		}
		retSheets = append(retSheets, &excelSheet)
	}
	retFile := &ExcelFile{}
	retFile.FileName = fileName
	retFile.Sheets = retSheets
	retFile.TotalRow = totalRow
	return retFile, nil
}

// checkExcelTitle 检测表头是否符合要求
func checkExcelTitle(header []string, checkTitles []string) (map[int]bool, error) {
	columnIsEmpty := make(map[int]bool)
	var titleNum int
	for i := range header {
		columnIsEmpty[i] = true
		if sliceutil.StringInSlice(checkTitles, header[i]) {
			titleNum++
		}
	}
	if titleNum != len(checkTitles) {
		var fullTitle string
		for _, title := range checkTitles {
			if fullTitle == "" {
				fullTitle = title
			} else {
				fullTitle += "," + title
			}
		}
		return nil, fmt.Errorf("表头不符合要求,表头应包含列:%s", fullTitle)
	}
	return columnIsEmpty, nil
}

// ProcessCSVFile 处理 .csv 文件
func ProcessCSVFile(fileName string, checkTitles []string, dstTitleMap map[string]string) (
	*ExcelFile, error) {
	// 读取文件内容
	data, err := os.ReadFile(fileName)
	if err != nil {
		logx.Errorf("ProcessCSVFile Failed to read file: %v", err)
		return nil, err
	}
	// 检测文件编码
	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest(data)
	if err != nil {
		return nil, err
	}

	// 打开文件进行解码
	file, err := os.Open(fileName)
	if err != nil {
		logx.Errorf("ProcessCSVFile Failed to open file: %v", err)
		return nil, err
	}
	defer file.Close()

	// 根据检测到的编码选择解码器
	var decoder *encoding.Decoder
	switch result.Charset {
	case common.CharsetISO88591:
		decoder = charmap.ISO8859_1.NewDecoder()
	case common.CharsetGBK:
		decoder = simplifiedchinese.GBK.NewDecoder()
	case common.CharsetGB18030:
		decoder = simplifiedchinese.GB18030.NewDecoder()
	default:
	}
	var reader *csv.Reader
	if decoder != nil {
		reader = csv.NewReader(transform.NewReader(file, decoder))
	} else {
		reader = csv.NewReader(file)
	}
	reader.FieldsPerRecord = -1 // 允许可变数量的字段
	reader.LazyQuotes = true
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}
	// 去除表头前后空格
	for i, v := range header {
		header[i] = strings.TrimSpace(v)
	}
	// 标记每一列是否有非空数据,并检测表头是否符合要求
	columnIsEmpty, err := checkExcelTitle(header, checkTitles)
	if err != nil {
		return nil, err
	}
	records := [][]string{}
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		records = append(records, record)
		for i, value := range record {
			if value != "" {
				columnIsEmpty[i] = false
			}
		}
	}
	// 根据数据是否为空更新 dstTitleMap
	for i, empty := range columnIsEmpty {
		if empty {
			delete(dstTitleMap, header[i])
		}
	}
	// csv 文件只有一个表单
	excelSheet := &ExcelSheet{SheetName: "csv"}
	// 把表头改为映射后的表头
	for _, title := range header {
		if dst, ok := dstTitleMap[title]; ok {
			excelSheet.Header = append(excelSheet.Header, dst)
		}
	}
	totalRow := 0
	// 处理每一行数据
	for _, record := range records {
		mappedData := mapCSVRowData(header, record, dstTitleMap)
		if len(mappedData) > 0 {
			rowData := make([]string, 0)
			for _, title := range excelSheet.Header {
				if value, ok := mappedData[title]; ok {
					rowData = append(rowData, value)
				}
			}
			totalRow++
			excelSheet.Rows = append(excelSheet.Rows, rowData)
		}
	}
	retSheets := make([]*ExcelSheet, 0)
	retSheets = append(retSheets, excelSheet)
	retFile := &ExcelFile{}
	retFile.FileName = fileName
	retFile.Sheets = retSheets
	retFile.TotalRow = totalRow
	return retFile, nil
}

// ProcessXLSFile 处理 .xls 文件
func ProcessXLSFile(fileName string, checkTitles []string, dstTitleMap map[string]string) (
	*ExcelFile, error) {
	workbook, err := xls.Open(fileName, "utf-8")
	if err != nil {
		logx.Errorf("ProcessXLSFile:%s, to open file: %v", fileName, err)
		return nil, err
	}
	var totalRow int
	retSheets := make([]*ExcelSheet, 0)
	for i := 0; i < workbook.NumSheets(); i++ {
		sheet := workbook.GetSheet(i)
		if sheet == nil {
			continue
		}
		excelSheet := &ExcelSheet{SheetName: sheet.Name}
		header, err := extractXLSHeader(sheet)
		if err != nil {
			return nil, err
		}
		// 标记每一列是否有非空数据,并检测表头是否符合要求
		columnIsEmpty, err := checkExcelTitle(header, checkTitles)
		if err != nil {
			return nil, err
		}
		// 遍历所有行，检测列是否有非空数据
		for j := 1; j <= int(sheet.MaxRow); j++ {
			row := sheet.Row(j)
			for i := 0; i < row.LastCol(); i++ {
				if row.Col(i) != "" {
					columnIsEmpty[i] = false
				}
			}
		}
		// 根据数据是否为空更新 dstTitleMap
		for i, empty := range columnIsEmpty {
			if empty {
				delete(dstTitleMap, header[i])
			}
		}
		for _, title := range header {
			if dst, ok := dstTitleMap[title]; ok {
				excelSheet.Header = append(excelSheet.Header, dst)
			}
		}
		// 处理每一行数据
		for j := 1; j <= int(sheet.MaxRow); j++ { // 跳过表头
			row := sheet.Row(j)
			mappedData := mapXLSRowData(header, row, dstTitleMap)
			if len(mappedData) > 0 {
				rowData := make([]string, 0)
				for _, title := range excelSheet.Header {
					if value, ok := mappedData[title]; ok {
						rowData = append(rowData, value)
					}
				}
				totalRow++
				excelSheet.Rows = append(excelSheet.Rows, rowData)
			}
		}
		retSheets = append(retSheets, excelSheet)
	}
	retFile := &ExcelFile{}
	retFile.FileName = fileName
	retFile.Sheets = retSheets
	retFile.TotalRow = totalRow
	return retFile, nil
}

// extractHeader 提取 .xlsx 文件的表头
func extractHeader(rows []*xlsx.Row) ([]string, error) {
	if len(rows) == 0 {
		return nil, fmt.Errorf("no rows found in sheet")
	}
	headerRow := rows[0]
	header := make([]string, len(headerRow.Cells))
	for i, cell := range headerRow.Cells {
		header[i] = strings.TrimSpace(cell.String())
	}
	return header, nil
}

// extractXLSHeader 提取 .xls 文件的表头
func extractXLSHeader(sheet *xls.WorkSheet) ([]string, error) {
	if sheet.MaxRow == 0 {
		return nil, fmt.Errorf("no rows found in sheet")
	}
	headerRow := sheet.Row(0)
	header := make([]string, headerRow.LastCol())
	for j := 0; j < headerRow.LastCol(); j++ {
		header[j] = strings.TrimSpace(headerRow.Col(j))
	}
	return header, nil
}

// mapRowData 根据映射关系提取列数据，过滤掉包含空单元格的行，并统一时间格式
func mapRowData(header []string, cells []*xlsx.Cell, dstTitleMap map[string]string) map[string]string {
	mappedData := make(map[string]string)
	for i, cell := range cells {
		// 跳过没有表头的列
		if i >= len(header) {
			continue
		}
		srcTitle := header[i]
		dstTitle, ok := dstTitleMap[srcTitle]
		if !ok {
			continue
		}
		value := cell.String()
		if strings.Contains(dstTitle, "时间") {
			parsedTime, err := timeutil.FormatDate(value, timeutil.DefaultTimeLayout)
			if err == nil {
				value = parsedTime
			}
		}
		mappedData[dstTitle] = value
	}
	return mappedData
}

// mapCSVRowData mapCSVRowData 根据映射关系提取并重命名列数据，过滤掉包含空单元格的行，并统一时间格式（CSV）
func mapCSVRowData(header []string, record []string, dstTitleMap map[string]string) map[string]string {
	mappedData := make(map[string]string)
	for i, value := range record {
		srcTitle := header[i]
		dstTitle, ok := dstTitleMap[srcTitle]
		if !ok {
			continue
		}
		if strings.Contains(dstTitle, "时间") {
			parsedTime, err := timeutil.FormatDate(value, timeutil.DefaultTimeLayout)
			if err == nil {
				value = parsedTime
			}
		}
		mappedData[dstTitle] = value
	}
	return mappedData
}

// 根据映射关系提取并重命名列数据，过滤掉包含空单元格的行，并统一时间格式（XLS）
func mapXLSRowData(header []string, row *xls.Row, dstTitleMap map[string]string) map[string]string {
	mappedData := make(map[string]string)
	for i := 0; i < row.LastCol(); i++ {
		value := row.Col(i)
		srcTitle := header[i]
		dstTitle, ok := dstTitleMap[srcTitle]
		if !ok {
			continue
		}
		if strings.Contains(dstTitle, "时间") {
			parsedTime, err := timeutil.FormatDate(value, timeutil.DefaultTimeLayout)
			if err == nil {
				value = parsedTime
			}
		}
		mappedData[dstTitle] = value
	}
	return mappedData
}
