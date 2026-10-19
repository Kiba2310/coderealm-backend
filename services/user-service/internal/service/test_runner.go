package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"BootDevClone/services/user-service/internal/domain"
	"BootDevClone/services/user-service/internal/infrastructure/piston"
)

func BuildChallengeTestCode(challengeSlug string, language string, userCode string, testCases []domain.DailyChallengeTestCase) string {
	tcJSON, _ := json.Marshal(testCases)

	switch language {
	case "python", "py":
		var invokeStr string
		switch challengeSlug {
		case "two-sum":
			invokeStr = "two_sum(*eval('[' + tc['input'] + ']'))"
		case "palindrome-check":
			invokeStr = "is_palindrome(eval(tc['input']))"
		case "balanced-parentheses":
			invokeStr = "is_valid(eval(tc['input']))"
		case "majority-element":
			invokeStr = "majority_element(eval(tc['input']))"
		case "roman-to-integer":
			invokeStr = "roman_to_int(eval(tc['input']))"
		case "dragon-fizzbuzz":
			invokeStr = "dragon_fizzbuzz(eval(tc['input']))"
			if strings.Contains(userCode, "dragon_fizbuzz") {
				invokeStr = "dragon_fizbuzz(eval(tc['input']))"
			}
		case "valid-anagram":
			invokeStr = "is_anagram(*eval('[' + tc['input'] + ']'))"
		case "longest-word-length":
			invokeStr = "longest_word_length(eval(tc['input']))"
		default:
			invokeStr = "solve(eval(tc['input']))"
		}

		return fmt.Sprintf(`%s

import json, sys

test_cases = json.loads('''%s''')
results = []

for i, tc in enumerate(test_cases):
    try:
        actual = %s
        exp_raw = tc["expected"]
        try:
            exp_val = json.loads(exp_raw)
        except Exception:
            exp_val = exp_raw
        
        passed = (actual == exp_val)
        if isinstance(actual, list) and isinstance(exp_val, list):
            passed = (sorted(actual) == sorted(exp_val)) if len(actual) == len(exp_val) else False
        
        results.append({
            "index": i + 1,
            "input": tc["input"],
            "expected": tc["expected"],
            "actual": json.dumps(actual, ensure_ascii=False) if isinstance(actual, (list, dict, bool)) else str(actual),
            "passed": passed
        })
    except Exception as e:
        results.append({
            "index": i + 1,
            "input": tc["input"],
            "expected": tc["expected"],
            "actual": "",
            "passed": False,
            "error": str(e)
        })

print("###TEST_RESULTS_START###")
print(json.dumps(results))
print("###TEST_RESULTS_END###")
`, userCode, string(tcJSON), invokeStr)

	case "javascript", "js", "node":
		var invokeStr string
		switch challengeSlug {
		case "two-sum":
			invokeStr = "twoSum(...JSON.parse('[' + tc.input + ']'))"
		case "palindrome-check":
			invokeStr = "isPalindrome(JSON.parse(tc.input))"
		case "balanced-parentheses":
			invokeStr = "isValid(JSON.parse(tc.input))"
		case "majority-element":
			invokeStr = "majorityElement(JSON.parse(tc.input))"
		case "roman-to-integer":
			invokeStr = "romanToInt(JSON.parse(tc.input))"
		case "dragon-fizzbuzz":
			invokeStr = "dragonFizzbuzz(JSON.parse(tc.input))"
		case "valid-anagram":
			invokeStr = "isAnagram(...JSON.parse('[' + tc.input + ']'))"
		case "longest-word-length":
			invokeStr = "longestWordLength(JSON.parse(tc.input))"
		default:
			invokeStr = "solve(JSON.parse(tc.input))"
		}

		return fmt.Sprintf(`%s

const testCases = JSON.parse('%s');
const results = [];

for (let i = 0; i < testCases.length; i++) {
    const tc = testCases[i];
    try {
        const actual = %s;
        let expVal;
        try { expVal = JSON.parse(tc.expected); } catch(e) { expVal = tc.expected; }
        
        let passed = JSON.stringify(actual) === JSON.stringify(expVal);
        if (Array.isArray(actual) && Array.isArray(expVal) && actual.length === expVal.length) {
            passed = JSON.stringify(actual.slice().sort()) === JSON.stringify(expVal.slice().sort());
        }
        
        results.push({
            index: i + 1,
            input: tc.input,
            expected: tc.expected,
            actual: typeof actual === 'object' ? JSON.stringify(actual) : String(actual),
            passed: passed
        });
    } catch(err) {
        results.push({
            index: i + 1,
            input: tc.input,
            expected: tc.expected,
            actual: '',
            passed: false,
            error: err.toString()
        });
    }
}

console.log("###TEST_RESULTS_START###");
console.log(JSON.stringify(results));
console.log("###TEST_RESULTS_END###");
`, userCode, string(tcJSON), invokeStr)

	case "php":
		var invokeStr string
		switch challengeSlug {
		case "two-sum":
			invokeStr = "twoSum(...json_decode('[' . $tc['input'] . ']', true))"
		case "palindrome-check":
			invokeStr = "isPalindrome(json_decode($tc['input'], true))"
		case "balanced-parentheses":
			invokeStr = "isValid(json_decode($tc['input'], true))"
		case "majority-element":
			invokeStr = "majorityElement(json_decode($tc['input'], true))"
		case "roman-to-integer":
			invokeStr = "romanToInt(json_decode($tc['input'], true))"
		case "dragon-fizzbuzz":
			invokeStr = "dragonFizzbuzz(json_decode($tc['input'], true))"
		case "valid-anagram":
			invokeStr = "isAnagram(...json_decode('[' . $tc['input'] . ']', true))"
		case "longest-word-length":
			invokeStr = "longestWordLength(json_decode($tc['input'], true))"
		default:
			invokeStr = "solve(json_decode($tc['input'], true))"
		}

		cleanUserCode := strings.TrimPrefix(strings.TrimSpace(userCode), "<?php")

		return fmt.Sprintf(`<?php
%s

$testCases = json_decode('%s', true);
$results = [];

foreach ($testCases as $i => $tc) {
    try {
        $actual = %s;
        $expVal = json_decode($tc['expected'], true);
        if ($expVal === null && $tc['expected'] !== 'null') {
            $expVal = $tc['expected'] === 'true' ? true : ($tc['expected'] === 'false' ? false : $tc['expected']);
        }
        $passed = (json_encode($actual) === json_encode($expVal));
        if (is_array($actual) && is_array($expVal) && count($actual) === count($expVal)) {
            $a = $actual; $b = $expVal; sort($a); sort($b);
            $passed = ($a === $b);
        }
        $results[] = [
            'index' => $i + 1,
            'input' => $tc['input'],
            'expected' => $tc['expected'],
            'actual' => is_array($actual) ? json_encode($actual) : (is_bool($actual) ? ($actual ? 'true' : 'false') : (string)$actual),
            'passed' => $passed
        ];
    } catch (Throwable $e) {
        $results[] = [
            'index' => $i + 1,
            'input' => $tc['input'],
            'expected' => $tc['expected'],
            'actual' => '',
            'passed' => false,
            'error' => $e->getMessage()
        ];
    }
}

echo "###TEST_RESULTS_START###\n";
echo json_encode($results) . "\n";
echo "###TEST_RESULTS_END###\n";
`, cleanUserCode, string(tcJSON), invokeStr)

	case "csharp", "c#", "cs":
		var testRunners string
		for i, tc := range testCases {
			var callSnippet string
			switch challengeSlug {
			case "two-sum":
				closeBracketIdx := strings.Index(tc.Input, "]")
				arrPart := tc.Input[:closeBracketIdx+1]
				targetPart := strings.TrimSpace(tc.Input[closeBracketIdx+1:])
				targetPart = strings.TrimPrefix(targetPart, ",")
				targetPart = strings.TrimSpace(targetPart)
				csharpArr := strings.Replace(strings.Replace(arrPart, "[", "{", 1), "]", "}", 1)
				callSnippet = fmt.Sprintf("Solution.TwoSum(new int[] %s, %s)", csharpArr, targetPart)
			case "palindrome-check":
				callSnippet = fmt.Sprintf("Solution.IsPalindrome(%s)", tc.Input)
			case "balanced-parentheses":
				callSnippet = fmt.Sprintf("Solution.IsValid(%s)", tc.Input)
			case "majority-element":
				csharpArr := strings.Replace(strings.Replace(tc.Input, "[", "{", 1), "]", "}", 1)
				callSnippet = fmt.Sprintf("Solution.MajorityElement(new int[] %s)", csharpArr)
			case "roman-to-integer":
				callSnippet = fmt.Sprintf("Solution.RomanToInt(%s)", tc.Input)
			case "dragon-fizzbuzz":
				callSnippet = fmt.Sprintf("Solution.DragonFizzbuzz(%s)", tc.Input)
			case "valid-anagram":
				parts := strings.Split(tc.Input, ",")
				callSnippet = fmt.Sprintf("Solution.IsAnagram(%s, %s)", parts[0], strings.TrimSpace(parts[1]))
			case "longest-word-length":
				callSnippet = fmt.Sprintf("Solution.LongestWordLength(%s)", tc.Input)
			default:
				callSnippet = "0"
			}

			testRunners += fmt.Sprintf(`
        try {
            var actual = %s;
            string actualStr = FormatVal(actual);
            string expStr = "%s".Trim();
            bool passed = actualStr.Replace(" ", "") == expStr.Replace(" ", "") || (actual != null && actual.ToString().ToLower() == expStr.ToLower());
            results.Add("{\"index\": %d, \"input\": \"%s\", \"expected\": \"%s\", \"actual\": \"" + actualStr.Replace("\"", "\\\"") + "\", \"passed\": " + (passed ? "true" : "false") + "}");
        } catch (Exception ex) {
            results.Add("{\"index\": %d, \"input\": \"%s\", \"expected\": \"%s\", \"actual\": \"\", \"passed\": false, \"error\": \"" + ex.Message.Replace("\"", "\\\"") + "\"}");
        }
`, callSnippet, strings.ReplaceAll(tc.Expected, `"`, `\"`), i+1, strings.ReplaceAll(tc.Input, `"`, `\"`), strings.ReplaceAll(tc.Expected, `"`, `\"`), i+1, strings.ReplaceAll(tc.Input, `"`, `\"`), strings.ReplaceAll(tc.Expected, `"`, `\"`))
		}

		return fmt.Sprintf(`using System;
using System.Collections.Generic;
using System.Text.RegularExpressions;

%s

class Program {
    static string FormatVal(object obj) {
        if (obj == null) return "null";
        if (obj is int[] arr) return "[" + string.Join(", ", arr) + "]";
        if (obj is List<string> list) {
            var items = new List<string>();
            foreach (var s in list) items.Add("\"" + s.Replace("\"", "\\\"") + "\"");
            return "[" + string.Join(", ", items) + "]";
        }
        if (obj is bool b) return b ? "true" : "false";
        return obj.ToString();
    }

    static void Main() {
        var results = new List<string>();
%s
        Console.WriteLine("###TEST_RESULTS_START###");
        Console.WriteLine("[" + string.Join(",", results) + "]");
        Console.WriteLine("###TEST_RESULTS_END###");
    }
}
`, userCode, testRunners)

	case "go", "golang":
		var testRunners string
		for i, tc := range testCases {
			var callSnippet string
			switch challengeSlug {
			case "two-sum":
				closeBracketIdx := strings.Index(tc.Input, "]")
				arrPart := tc.Input[:closeBracketIdx+1]
				targetPart := strings.TrimSpace(tc.Input[closeBracketIdx+1:])
				targetPart = strings.TrimPrefix(targetPart, ",")
				targetPart = strings.TrimSpace(targetPart)
				goArr := strings.Replace(strings.Replace(arrPart, "[", "{", 1), "]", "}", 1)
				callSnippet = fmt.Sprintf("TwoSum([]int%s, %s)", goArr, targetPart)
			case "palindrome-check":
				callSnippet = fmt.Sprintf("IsPalindrome(%s)", tc.Input)
			case "balanced-parentheses":
				callSnippet = fmt.Sprintf("IsValid(%s)", tc.Input)
			case "majority-element":
				goArr := strings.Replace(strings.Replace(tc.Input, "[", "{", 1), "]", "}", 1)
				callSnippet = fmt.Sprintf("MajorityElement([]int%s)", goArr)
			case "roman-to-integer":
				callSnippet = fmt.Sprintf("RomanToInt(%s)", tc.Input)
			case "dragon-fizzbuzz":
				callSnippet = fmt.Sprintf("DragonFizzbuzz(%s)", tc.Input)
			case "valid-anagram":
				parts := strings.Split(tc.Input, ",")
				callSnippet = fmt.Sprintf("IsAnagram(%s, %s)", parts[0], strings.TrimSpace(parts[1]))
			case "longest-word-length":
				callSnippet = fmt.Sprintf("LongestWordLength(%s)", tc.Input)
			default:
				callSnippet = "nil"
			}

			testRunners += fmt.Sprintf(`
	{
		actual := %s
		actBytes, _ := json.Marshal(actual)
		actStr := string(actBytes)
		expStr := %q
		passed := (actStr == expStr) || (fmt.Sprintf("%%v", actual) == expStr)
		results = append(results, TestResult{
			Index: %d,
			Input: %q,
			Expected: expStr,
			Actual: actStr,
			Passed: passed,
		})
	}
`, callSnippet, tc.Expected, i+1, tc.Input)
		}

		cleanUserCode := strings.TrimPrefix(strings.TrimSpace(userCode), "package main")

		return fmt.Sprintf(`package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

%s

func main() {
	type TestResult struct {
		Index    int    `+"`json:\"index\"`"+`
		Input    string `+"`json:\"input\"`"+`
		Expected string `+"`json:\"expected\"`"+`
		Actual   string `+"`json:\"actual\"`"+`
		Passed   bool   `+"`json:\"passed\"`"+`
		Error    string `+"`json:\"error,omitempty\"`"+`
	}
	var results []TestResult

%s

	data, _ := json.Marshal(results)
	fmt.Println("###TEST_RESULTS_START###")
	fmt.Println(string(data))
	fmt.Println("###TEST_RESULTS_END###")
}
`, cleanUserCode, testRunners)

	default:
		return userCode
	}
}

func ParseTestResults(output string, compileOutput string) ([]domain.TestResultItem, bool, string) {
	if compileOutput != "" {
		return nil, false, compileOutput
	}

	startTag := "###TEST_RESULTS_START###"
	endTag := "###TEST_RESULTS_END###"

	startIdx := strings.Index(output, startTag)
	endIdx := strings.Index(output, endTag)

	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		return nil, false, output
	}

	jsonStr := strings.TrimSpace(output[startIdx+len(startTag) : endIdx])
	var items []domain.TestResultItem
	if err := json.Unmarshal([]byte(jsonStr), &items); err != nil {
		return nil, false, "Teszt eredmények feldolgozási hiba: " + err.Error() + "\n" + output
	}

	allPassed := true
	for _, item := range items {
		if !item.Passed {
			allPassed = false
			break
		}
	}

	return items, allPassed, ""
}

func ExecuteChallengeTests(pistonClient *piston.Client, challengeSlug string, language string, userCode string, testCases []domain.DailyChallengeTestCase) (*domain.RunChallengeResponse, error) {
	wrappedCode := BuildChallengeTestCode(challengeSlug, language, userCode, testCases)
	resp, err := pistonClient.ExecuteCode(language, wrappedCode)
	if err != nil {
		return nil, fmt.Errorf("kódfuttatási hiba: %w", err)
	}

	var compileErr string
	if resp.Compile.Stderr != "" {
		compileErr = resp.Compile.Stderr
	} else if resp.Compile.Output != "" && resp.Compile.Code != 0 {
		compileErr = resp.Compile.Output
	}

	rawOutput := resp.Run.Stdout
	if resp.Run.Stderr != "" {
		rawOutput += "\n[STDERR]: " + resp.Run.Stderr
	}

	testItems, allPassed, parseErr := ParseTestResults(rawOutput, compileErr)

	cleanConsoleOutput := rawOutput
	if startIdx := strings.Index(cleanConsoleOutput, "###TEST_RESULTS_START###"); startIdx != -1 {
		cleanConsoleOutput = strings.TrimSpace(cleanConsoleOutput[:startIdx])
	}

	passedCount := 0
	for _, item := range testItems {
		if item.Passed {
			passedCount++
		}
	}

	return &domain.RunChallengeResponse{
		Success:       allPassed && len(testItems) > 0,
		TestResults:   testItems,
		ConsoleOutput: cleanConsoleOutput,
		TotalPassed:   passedCount,
		TotalTests:    len(testCases),
		Error:         parseErr,
	}, nil
}
