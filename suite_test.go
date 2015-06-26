package godog

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cucumber/gherkin-go"
)

func SuiteContext(s Suite) {
	c := &suiteContext{}

	s.BeforeScenario(c.ResetBeforeEachScenario)

	s.Step(`^a feature path "([^"]*)"$`, c.featurePath)
	s.Step(`^I parse features$`, c.parseFeatures)
	s.Step(`^I'm listening to suite events$`, c.iAmListeningToSuiteEvents)
	s.Step(`^I run feature suite$`, c.iRunFeatureSuite)
	s.Step(`^a feature "([^"]*)" file:$`, c.aFeatureFile)
	s.Step(`^the suite should have (passed|failed)$`, c.theSuiteShouldHave)

	s.Step(`^I should have ([\d]+) features? files?:$`, c.iShouldHaveNumFeatureFiles)
	s.Step(`^I should have ([\d]+) scenarios? registered$`, c.numScenariosRegistered)
	s.Step(`^there (was|were) ([\d]+) "([^"]*)" events? fired$`, c.thereWereNumEventsFired)
	s.Step(`^there was event triggered before scenario "([^"]*)"$`, c.thereWasEventTriggeredBeforeScenario)
	s.Step(`^these events had to be fired for a number of times:$`, c.theseEventsHadToBeFiredForNumberOfTimes)

	s.Step(`^a failing step`, c.aFailingStep)
	s.Step(`^this step should fail`, c.aFailingStep)
	s.Step(`^the following steps? should be (passed|failed|skipped|undefined|pending):`, c.followingStepsShouldHave)

	// lt
	s.Step(`^savybių aplankas "([^"]*)"$`, c.featurePath)
	s.Step(`^aš išskaitau savybes$`, c.parseFeatures)
	s.Step(`^aš turėčiau turėti ([\d]+) savybių failus:$`, c.iShouldHaveNumFeatureFiles)

	s.Step(`^pending step$`, func(...*Arg) error {
		return ErrPending
	})
}

type firedEvent struct {
	name string
	args []interface{}
}

type suiteContext struct {
	testedSuite *suite
	events      []*firedEvent
	fmt         *testFormatter
}

func (s *suiteContext) ResetBeforeEachScenario(interface{}) {
	// reset whole suite with the state
	s.fmt = &testFormatter{}
	s.testedSuite = &suite{fmt: s.fmt}
	// our tested suite will have the same context registered
	SuiteContext(s.testedSuite)
	// reset all fired events
	s.events = []*firedEvent{}
}

func (s *suiteContext) followingStepsShouldHave(status string, steps *gherkin.DocString) error {
	var expected = strings.Split(steps.Content, "\n")
	var actual, unmatched []string
	var matched []int

	switch status {
	case "passed":
		for _, st := range s.fmt.passed {
			actual = append(actual, st.step.Text)
		}
	case "failed":
		for _, st := range s.fmt.failed {
			actual = append(actual, st.step.Text)
		}
	case "skipped":
		for _, st := range s.fmt.skipped {
			actual = append(actual, st.step.Text)
		}
	case "undefined":
		for _, st := range s.fmt.undefined {
			actual = append(actual, st.step.Text)
		}
	case "pending":
		for _, st := range s.fmt.pending {
			actual = append(actual, st.step.Text)
		}
	default:
		return fmt.Errorf("unexpected step status wanted: %s", status)
	}

	if len(expected) > len(actual) {
		return fmt.Errorf("number of expected %s steps: %d is less than actual %s steps: %d", status, len(expected), status, len(actual))
	}

	for _, a := range actual {
		for i, e := range expected {
			if a == e {
				matched = append(matched, i)
				break
			}
		}
	}

	if len(matched) == len(expected) {
		return nil
	}

	for i, s := range expected {
		var found bool
		for _, m := range matched {
			if i == m {
				found = true
			}
		}
		if !found {
			unmatched = append(unmatched, s)
		}
	}

	return fmt.Errorf("the steps: %s - is not %s", strings.Join(unmatched, ", "), status)
}

func (s *suiteContext) iAmListeningToSuiteEvents() error {
	s.testedSuite.BeforeSuite(func() {
		s.events = append(s.events, &firedEvent{"BeforeSuite", []interface{}{}})
	})
	s.testedSuite.AfterSuite(func() {
		s.events = append(s.events, &firedEvent{"AfterSuite", []interface{}{}})
	})
	s.testedSuite.BeforeScenario(func(scenario interface{}) {
		s.events = append(s.events, &firedEvent{"BeforeScenario", []interface{}{scenario}})
	})
	s.testedSuite.AfterScenario(func(scenario interface{}, err error) {
		s.events = append(s.events, &firedEvent{"AfterScenario", []interface{}{scenario, err}})
	})
	s.testedSuite.BeforeStep(func(step *gherkin.Step) {
		s.events = append(s.events, &firedEvent{"BeforeStep", []interface{}{step}})
	})
	s.testedSuite.AfterStep(func(step *gherkin.Step, err error) {
		s.events = append(s.events, &firedEvent{"AfterStep", []interface{}{step, err}})
	})
	return nil
}

func (s *suiteContext) aFailingStep() error {
	return fmt.Errorf("intentional failure")
}

// parse a given feature file body as a feature
func (s *suiteContext) aFeatureFile(name string, body *gherkin.DocString) error {
	ft, err := gherkin.ParseFeature(strings.NewReader(body.Content))
	s.testedSuite.features = append(s.testedSuite.features, &feature{Feature: ft, Path: name})
	return err
}

func (s *suiteContext) featurePath(path string) error {
	s.testedSuite.paths = append(s.testedSuite.paths, path)
	return nil
}

func (s *suiteContext) parseFeatures() error {
	return s.testedSuite.parseFeatures()
}

func (s *suiteContext) theSuiteShouldHave(state string) error {
	if s.testedSuite.failed && state == "passed" {
		return fmt.Errorf("the feature suite has failed")
	}
	if !s.testedSuite.failed && state == "failed" {
		return fmt.Errorf("the feature suite has passed")
	}
	return nil
}

func (s *suiteContext) iShouldHaveNumFeatureFiles(num int, files *gherkin.DocString) error {
	if len(s.testedSuite.features) != num {
		return fmt.Errorf("expected %d features to be parsed, but have %d", num, len(s.testedSuite.features))
	}
	expected := strings.Split(files.Content, "\n")
	var actual []string
	for _, ft := range s.testedSuite.features {
		actual = append(actual, ft.Path)
	}
	if len(expected) != len(actual) {
		return fmt.Errorf("expected %d feature paths to be parsed, but have %d", len(expected), len(actual))
	}
	for i := 0; i < len(expected); i++ {
		if expected[i] != actual[i] {
			return fmt.Errorf(`expected feature path "%s" at position: %d, does not match actual "%s"`, expected[i], i, actual[i])
		}
	}
	return nil
}

func (s *suiteContext) iRunFeatureSuite() error {
	if err := s.parseFeatures(); err != nil {
		return err
	}
	s.testedSuite.run()
	return nil
}

func (s *suiteContext) numScenariosRegistered(expected int) (err error) {
	var num int
	for _, ft := range s.testedSuite.features {
		num += len(ft.ScenarioDefinitions)
	}
	if num != expected {
		err = fmt.Errorf("expected %d scenarios to be registered, but got %d", expected, num)
	}
	return
}

func (s *suiteContext) thereWereNumEventsFired(_ string, expected int, typ string) error {
	var num int
	for _, event := range s.events {
		if event.name == typ {
			num++
		}
	}
	if num != expected {
		return fmt.Errorf("expected %d %s events to be fired, but got %d", expected, typ, num)
	}
	return nil
}

func (s *suiteContext) thereWasEventTriggeredBeforeScenario(expected string) error {
	var found []string
	for _, event := range s.events {
		if event.name != "BeforeScenario" {
			continue
		}

		var name string
		switch t := event.args[0].(type) {
		case *gherkin.Scenario:
			name = t.Name
		case *gherkin.ScenarioOutline:
			name = t.Name
		}
		if name == expected {
			return nil
		}

		found = append(found, name)
	}

	if len(found) == 0 {
		return fmt.Errorf("before scenario event was never triggered or listened")
	}

	return fmt.Errorf(`expected "%s" scenario, but got these fired %s`, expected, `"`+strings.Join(found, `", "`)+`"`)
}

func (s *suiteContext) theseEventsHadToBeFiredForNumberOfTimes(tbl *gherkin.DataTable) error {
	if len(tbl.Rows[0].Cells) != 2 {
		return fmt.Errorf("expected two columns for event table row, got: %d", len(tbl.Rows[0].Cells))
	}

	for _, row := range tbl.Rows {
		num, err := strconv.ParseInt(row.Cells[1].Value, 10, 0)
		if err != nil {
			return err
		}
		if err := s.thereWereNumEventsFired("", int(num), row.Cells[0].Value); err != nil {
			return err
		}
	}
	return nil
}
