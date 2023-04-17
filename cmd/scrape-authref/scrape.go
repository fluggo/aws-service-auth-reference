package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

const (
	startPage       = "https://docs.aws.amazon.com/service-authorization/latest/reference/reference_policies_actions-resources-contextkeys.html"
	testActionsPage = "https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazonec2.html"
)

var (
	spaceReplacer = regexp.MustCompile(`\s{2,}`)
)

func mustParseSelector(sel string) cascadia.Sel {
	result, err := cascadia.Parse(sel)

	if err != nil {
		panic(err)
	}

	return result
}

func gatherText(node *html.Node, recursive bool) string {
	result := ""

	for childNode := node.FirstChild; childNode != nil; childNode = childNode.NextSibling {
		if childNode.Type == html.TextNode {
			result += childNode.Data
		} else if recursive {
			result += gatherText(childNode, true)
		}
	}

	return spaceReplacer.ReplaceAllLiteralString(strings.TrimSpace(result), " ")
}

func renderToString(node *html.Node) string {
	if node == nil {
		return ""
	}

	var buf bytes.Buffer
	html.Render(&buf, node)
	return buf.String()
}

func fetchHtml(url string) (*html.Node, error) {
	resp, err := http.Get(url)

	if err != nil {
		return nil, fmt.Errorf("HTTP GET: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP GET: status code %v", resp.StatusCode)
	}

	node, err := html.Parse(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	return node, nil
}

type topic struct {
	name string
	url  *url.URL
}

func getAttrValue(node *html.Node, name string) string {
	for _, v := range node.Attr {
		if v.Key == name {
			return v.Val
		}
	}

	return ""
}

func parseTopics() ([]topic, error) {
	node, err := fetchHtml(startPage)

	if err != nil {
		return nil, fmt.Errorf("parseTopics: %w", err)
	}

	// Not fully documented in cascadia, but it has these additional text selectors:
	//
	//	:contains("str")		Selects nodes that contain the given text when all descendant text nodes are combined
	//	:containsOwn("str")		Selects nodes that contain the given text when all child text nodes are combined
	//	:matches(^[a-z]$)		Selects nodes that match the given regex when all descendant text nodes are combined
	//	:matchesOwn(^[a-z]$)	Selects nodes that match the given regex when all child text nodes are combined
	//	:has(selector)			Selects nodes that contain descendant nodes that match the given selector
	//	:haschild(selector)		Selects nodes that contain child nodes that match the given selector
	//	:input					Selects any input element (input, select, textarea, or button)
	//	[attr#=(^[a-z]$)]		Selects elements with attributes that match the given regex
	//
	// Additionally, it implements all the tree-structural pseudo-classes found here:
	//	https://developer.mozilla.org/en-US/docs/Web/CSS/Pseudo-classes#tree-structural_pseudo-classes

	topicsListSelector := mustParseSelector(`h6:matchesOwn(^\s*Topics\s*$) + ul`)
	topicsListNode := cascadia.Query(node, topicsListSelector)

	if topicsListNode == nil {
		return nil, fmt.Errorf("get topics: could not find topics")
	}

	result := make([]topic, 0, 20)
	baseUrl, err := url.Parse(startPage)

	if err != nil {
		panic(err)
	}

	topicsSelector := mustParseSelector(`li > a`)
	topicsNodes := cascadia.QueryAll(topicsListNode, topicsSelector)

	for _, aNode := range topicsNodes {
		partialHref := getAttrValue(aNode, "href")
		title := aNode.FirstChild.Data

		if partialHref == "" {
			return nil, fmt.Errorf("get topics: could not find topic <a> href")
		}

		newUrl, err := baseUrl.Parse(partialHref)

		if err != nil {
			return nil, fmt.Errorf("get topics: parse URL %s: %w", partialHref, err)
		}

		result = append(result, topic{name: title, url: newUrl})
	}

	return result, nil
}

type ServiceAuthorizationReference struct {
	Name              string          `json:"name"`
	ServicePrefix     string          `json:"servicePrefix"`
	AuthReferenceHref string          `json:"authReferenceHref"`
	ApiReferenceHref  string          `json:"apiReferenceHref,omitempty"`
	Actions           []*Action       `json:"actions"`
	ResourceTypes     []*ResourceType `json:"resourceTypes"`
	ConditionKeys     []*ConditionKey `json:"conditionKeys"`
}

type ActionResourceType struct {
	ResourceType     string   `json:"resourceType"`
	Required         bool     `json:"required"`
	ConditionKeys    []string `json:"conditionKeys"`
	DependentActions []string `json:"dependentActions"`
}

type Action struct {
	Name           string               `json:"name"`
	PermissionOnly bool                 `json:"permissionOnly"`
	ReferenceHref  string               `json:"referenceHref,omitempty"`
	Description    string               `json:"description"`
	AccessLevel    string               `json:"accessLevel"`
	ResourceTypes  []ActionResourceType `json:"resourceTypes"`
}

func parseAPIReferenceHref(page *html.Node) string {
	apiReferenceLink := mustParseSelector(`#main-col-body a[href]:containsOwn("API operations available for")`)

	if apiReferenceNode := cascadia.Query(page, apiReferenceLink); apiReferenceNode != nil {
		return getAttrValue(apiReferenceNode, "href")
	} else {
		return ""
	}
}

func parseServicePrefix(page *html.Node) string {
	servicePrefixSelector := mustParseSelector(`#main-col-body > p:containsOwn("service prefix:") > code[class*="code"]`)
	servicePrefixNode := cascadia.Query(page, servicePrefixSelector)

	return servicePrefixNode.FirstChild.Data
}

func parseActionsTable(page *html.Node) ([]*Action, error) {
	actionTableSelector := mustParseSelector(`h2:containsOwn("Actions defined by") ~ div[class*="table-container"] table`)
	actionTableNode := cascadia.Query(page, actionTableSelector)

	rowSelector := mustParseSelector(`tr`)
	rowNodes := cascadia.QueryAll(actionTableNode, rowSelector)

	cellSelector := mustParseSelector(`td`)
	aHrefSelector := mustParseSelector(`a[href]`)
	pSelector := mustParseSelector(`p`)
	actions := make([]*Action, 0)
	var action *Action
	var nextActionRow, nextDescriptionRow int

	for row := 1; row < len(rowNodes); row++ {
		rowNode := rowNodes[row]
		rowCellNodes := cascadia.QueryAll(rowNode, cellSelector)

		if action == nil || row == nextActionRow {
			action = &Action{}
			actions = append(actions, action)

			if len(rowCellNodes) != 6 {
				return nil, fmt.Errorf("first row of action table entry has %d cells (expected 6): %#v", len(rowCellNodes), renderToString(rowNode))
			}

			actionRowspan := 1

			if rowspanValue := getAttrValue(rowCellNodes[0], "rowspan"); rowspanValue != "" {
				if v, err := strconv.Atoi(rowspanValue); err == nil {
					actionRowspan = v
				}
			}

			nextActionRow = row + actionRowspan
			nextDescriptionRow = row
			actionNameRaw := gatherText(rowCellNodes[0], true)
			actionNameSubstrings := strings.SplitN(actionNameRaw, " ", 2)

			if actionNameNode := cascadia.Query(rowCellNodes[0], aHrefSelector); actionNameNode != nil {
				action.Name = gatherText(actionNameNode, true)
				action.ReferenceHref = getAttrValue(actionNameNode, "href")
			} else {
				action.Name = actionNameSubstrings[0]
			}

			if strings.Contains(actionNameRaw, "[permission only]") {
				action.PermissionOnly = true
			}

			action.ResourceTypes = make([]ActionResourceType, 0)
		}

		if row == nextDescriptionRow {
			descriptionRowspan := 1
			descriptionCellNode := rowCellNodes[len(rowCellNodes)-5]

			if rowspanValue := getAttrValue(descriptionCellNode, "rowspan"); rowspanValue != "" {
				if v, err := strconv.Atoi(rowspanValue); err == nil {
					descriptionRowspan = v
				}
			}

			nextDescriptionRow = row + descriptionRowspan

			// For now, we only take the first description we find; the "SCENARIO" blocks in the EC2 documentation aren't interesting to us
			if action.Description != "" {
				row = nextActionRow - 1
				continue
			}

			action.Description = gatherText(descriptionCellNode, true)

			accessLevelNode := rowCellNodes[len(rowCellNodes)-4]
			action.AccessLevel = gatherText(accessLevelNode, true)
		}

		resourceType := ActionResourceType{}

		resourceTypeField := gatherText(rowCellNodes[len(rowCellNodes)-3], true)
		resourceType.ResourceType = strings.TrimSuffix(resourceTypeField, "*")
		resourceType.Required = strings.HasSuffix(resourceTypeField, "*")

		conditionKeyNodes := cascadia.QueryAll(rowCellNodes[len(rowCellNodes)-2], pSelector)
		resourceType.ConditionKeys = make([]string, len(conditionKeyNodes))

		for k, conditionKeyNode := range conditionKeyNodes {
			resourceType.ConditionKeys[k] = gatherText(conditionKeyNode, true)
		}

		dependentActionNodes := cascadia.QueryAll(rowCellNodes[len(rowCellNodes)-1], pSelector)
		resourceType.DependentActions = make([]string, len(dependentActionNodes))

		for k, dependentActionNode := range dependentActionNodes {
			resourceType.DependentActions[k] = gatherText(dependentActionNode, true)
		}

		if resourceType.ResourceType != "" {
			action.ResourceTypes = append(action.ResourceTypes, resourceType)
		}
	}

	return actions, nil
}

type ResourceType struct {
	Name          string   `json:"name"`
	ReferenceHref string   `json:"referenceHref,omitempty"`
	ArnPattern    string   `json:"arnPattern"`
	ConditionKeys []string `json:"conditionKeys"`
}

func parseResourceTypesTable(page *html.Node) []*ResourceType {
	rtTableSelector := mustParseSelector(`h2:containsOwn("Resource types defined by") + p + div[class*="table-container"] table`)
	rtTableNode := cascadia.Query(page, rtTableSelector)

	if rtTableNode == nil {
		return make([]*ResourceType, 0)
	}

	rowSelector := mustParseSelector(`tr`)
	rowNodes := cascadia.QueryAll(rtTableNode, rowSelector)

	cellSelector := mustParseSelector(`td`)
	aHrefSelector := mustParseSelector(`a[href]`)
	pSelector := mustParseSelector(`p`)
	resourceTypes := make([]*ResourceType, 0)
	var resourceType *ResourceType

	for row := 1; row < len(rowNodes); row++ {
		rowNode := rowNodes[row]
		rowCellNodes := cascadia.QueryAll(rowNode, cellSelector)

		resourceType = &ResourceType{}
		resourceTypes = append(resourceTypes, resourceType)

		if len(rowCellNodes) != 3 {
			panic(fmt.Errorf("first row of resource table entry has %d cells (expected 3): %#v", len(rowCellNodes), renderToString(rowNode)))
		}

		resourceType.Name = gatherText(rowCellNodes[0], true)

		if resourceTypeRefLink := cascadia.Query(rowCellNodes[0], aHrefSelector); resourceTypeRefLink != nil {
			resourceType.ReferenceHref = getAttrValue(resourceTypeRefLink, "href")
		}

		resourceType.ArnPattern = gatherText(rowCellNodes[1], true)

		conditionKeyNodes := cascadia.QueryAll(rowCellNodes[2], pSelector)
		resourceType.ConditionKeys = make([]string, len(conditionKeyNodes))

		for k, conditionKeyNode := range conditionKeyNodes {
			resourceType.ConditionKeys[k] = gatherText(conditionKeyNode, true)
		}
	}

	return resourceTypes
}

type ConditionKey struct {
	Name          string `json:"name"`
	ReferenceHref string `json:"referenceHref,omitempty"`
	Description   string `json:"description"`
	Type          string `json:"type"`
}

func parseConditionKeyTable(page *html.Node) []*ConditionKey {
	ckTableSelector := mustParseSelector(`h2:containsOwn("Condition keys for") + p + p + div[class*="table-container"] table`)
	ckTableNode := cascadia.Query(page, ckTableSelector)

	if ckTableNode == nil {
		return make([]*ConditionKey, 0)
	}

	rowSelector := mustParseSelector(`tr`)
	rowNodes := cascadia.QueryAll(ckTableNode, rowSelector)

	cellSelector := mustParseSelector(`td`)
	aHrefSelector := mustParseSelector(`a[href]`)
	// pSelector := mustParseSelector(`p`)
	conditionKeys := make([]*ConditionKey, 0)
	var conditionKey *ConditionKey

	for row := 1; row < len(rowNodes); row++ {
		rowNode := rowNodes[row]
		rowCellNodes := cascadia.QueryAll(rowNode, cellSelector)

		conditionKey = &ConditionKey{}
		conditionKeys = append(conditionKeys, conditionKey)

		if len(rowCellNodes) != 3 {
			fmt.Printf("%s\n", renderToString(rowNode))
			panic(fmt.Errorf("first row of condition key entry has %d cells (expected 3)", len(rowCellNodes)))
		}

		conditionKey.Name = gatherText(rowCellNodes[0], true)

		if refLink := cascadia.Query(rowCellNodes[0], aHrefSelector); refLink != nil {
			conditionKey.ReferenceHref = getAttrValue(refLink, "href")
		}

		conditionKey.Description = gatherText(rowCellNodes[1], true)
		conditionKey.Type = gatherText(rowCellNodes[2], true)
	}

	return conditionKeys
}

func main() {
	topics, err := parseTopics()

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse topics page: %v\n", err)
		os.Exit(1)
	}

	authRefs := make([]*ServiceAuthorizationReference, 0)

	for _, topic := range topics {
		page, err := fetchHtml(topic.url.String())

		if err != nil {
			fmt.Fprintf(os.Stderr, "topic %#v: %v\n", topic.name, err)
			os.Exit(1)
		}

		authRef := &ServiceAuthorizationReference{Name: topic.name, AuthReferenceHref: topic.url.String()}
		authRefs = append(authRefs, authRef)

		if actions, err := parseActionsTable(page); err != nil {
			fmt.Fprintf(os.Stderr, "topic %#v: actions table: %v\n", topic.name, err)
			os.Exit(1)
		} else {
			authRef.Actions = actions
		}

		authRef.ConditionKeys = parseConditionKeyTable(page)
		authRef.ResourceTypes = parseResourceTypesTable(page)
		authRef.ApiReferenceHref = parseAPIReferenceHref(page)
		authRef.ServicePrefix = parseServicePrefix(page)
	}

	indentedFile, err := os.Create("service-auth.json")

	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open output file: %v\n", err)
		os.Exit(1)
	}

	encoder := json.NewEncoder(indentedFile)
	encoder.SetIndent("", "  ")

	encoder.Encode(authRefs)

	if err := indentedFile.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "could not close output file: %v\n", err)
		os.Exit(1)
	}
}
