package excelutil

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/extrame/xls"
	"github.com/gogs/chardet"
	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func OpenExFile(fileName string) (*ExcelFile, error) {
	var retFile ExcelFile
	retSheets := make([]*ExcelSheet, 0)
	totalRow := 0
	//打开xlsx
	if strings.HasSuffix(fileName, ".xlsx") {
		f, err := excelize.OpenFile(fileName)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		sheetList := f.GetSheetList()
		for _, name := range sheetList {
			var retSheet ExcelSheet
			rows, err := f.GetRows(name)
			if err != nil {
				return nil, err
			}
			totalRow += len(rows)
			retSheet.SheetName = name
			retSheet.Rows = rows
			retSheets = append(retSheets, &retSheet)
		}
	}
	//打开xls
	if strings.HasSuffix(fileName, ".xls") {
		dealXls(fileName)
	}

	//打开csv
	if strings.HasSuffix(fileName, ".csv") {
		csvFile, err := dealCSV(fileName)
		if err != nil {
			return nil, err
		}
		var retSheet ExcelSheet
		retSheet.SheetName = "csv"
		retSheet.Rows = csvFile
		totalRow = len(csvFile)
		retSheets = append(retSheets, &retSheet)
	}
	retFile.Sheets = retSheets
	retFile.TotalRow = totalRow
	retFile.FileName = fileName
	return &retFile, nil
}

func dealCSV(fileName string) ([][]string, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		logx.Errorf("Failed to read file: %v", err)
		return nil, err
	}

	// 检测文件编码
	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest(data)
	if err != nil {
		log.Fatalf("Failed to detect encoding: %v", err)
	}

	// 打开文件进行解码
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// 根据检测到的编码选择解码器
	var decoder *encoding.Decoder
	switch result.Charset {
	case "UTF-8":
		decoder = nil // UTF-8 不需要解码器
	case "ISO-8859-1":
		decoder = charmap.ISO8859_1.NewDecoder()
	case "GBK":
		decoder = simplifiedchinese.GBK.NewDecoder()
	case "GB18030":
		decoder = simplifiedchinese.GB18030.NewDecoder()
	default:
		return nil, errors.New(fmt.Sprintf("Unsupported encoding: %s", result.Charset))
	}

	// 创建 CSV reader
	var reader *csv.Reader
	if decoder != nil {
		reader = csv.NewReader(transform.NewReader(file, decoder))
	} else {
		reader = csv.NewReader(file)
	}
	// 读取所有记录
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV file: %v", err)
	}
	return records, nil
}

func dealXls(fileName string) ([]*ExcelSheet, error) {
	f, err := xls.Open(fileName, "utf-8")
	if err != nil {
		return nil, err
	}
	var totalRow int
	var retSheets []*ExcelSheet
	for i := 0; i < f.NumSheets(); i++ {
		var retSheet ExcelSheet
		sheet := f.GetSheet(i)
		totalRow += int(sheet.MaxRow) - 1
		retRows := make([][]string, 0, 64)
		for j := 0; j < int(sheet.MaxRow); j++ {
			retRow := make([]string, 0, 64)
			row := sheet.Row(j)
			if row != nil {
				for k := row.FirstCol(); k <= row.LastCol(); k++ {
					col := row.Col(k)
					retRow = append(retRow, col)
				}
				retRows = append(retRows, retRow)
			}
		}
		retSheet.Rows = retRows
		retSheet.SheetName = sheet.Name
		retSheets = append(retSheets, &retSheet)
	}
	return retSheets, nil
}
func IsXlsx(fileName string) bool {
	if strings.HasSuffix(fileName, ".xlsx") || strings.HasSuffix(fileName, ".xls") {
		return true
	}
	return false
}

func IsExcel(fileName string) bool {
	if strings.HasSuffix(fileName, ".xlsx") || strings.HasSuffix(fileName, ".xls") || strings.HasSuffix(fileName, ".csv") {
		return true
	}
	return false
}

func GetExcelMakeCSV(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	lastName := "output.csv"
	// 创建一个新文件
	fileName := fmt.Sprintf("%v"+lastName, time.Now().Unix())
	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 将Excel文件写入新文件
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}

	return fileName, nil
}

func GetExcel(url, inputName string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	lastName := "output.xlsx"
	if strings.HasSuffix(inputName, ".xls") {
		lastName = "output.xls"
	}
	if strings.HasSuffix(inputName, ".csv") {
		lastName = "output.csv"
	}
	// 创建一个新文件
	fileName := fmt.Sprintf("%v"+lastName, time.Now().Unix())
	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 将Excel文件写入新文件
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}

	return fileName, nil
}

func WriteDataToExcel(dataMap map[string][][]string, fileName string, tittle []string) error {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	for sheet, data := range dataMap {
		// Create a new sheet for each key
		index, _ := f.NewSheet(sheet)

		// Write tittle to the sheet
		for i, cell := range tittle {
			colName, _ := excelize.ColumnNumberToName(i + 1)
			// Excel coordinates start from 1, not 0
			_ = f.SetCellValue(sheet, fmt.Sprintf("%s%d", colName, 1), cell)
		}

		// Write data to the sheet
		for i, row := range data {
			for j, cell := range row {
				colName, _ := excelize.ColumnNumberToName(j + 1)

				// Excel coordinates start from 1, not 0
				_ = f.SetCellValue(sheet, fmt.Sprintf("%s%d", colName, i+2), cell)
			}
		}

		// Set the created sheet as the active sheet
		f.SetActiveSheet(index)
	}
	// Save the file
	if err := f.SaveAs(fileName); err != nil {
		return err
	}

	return nil
}

func MacStrToInt(macStr string) (int64, error) {
	macStr = strings.Replace(macStr, ":", "", -1)
	macStr = strings.Replace(macStr, "-", "", -1)

	if len(macStr) != 12 {
		return 0, errors.New("MAC字符串要为12字节")
	}

	return strconv.ParseInt(macStr, 16, 64)
}
