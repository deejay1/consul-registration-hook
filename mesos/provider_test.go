package mesos

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestIfReturnsServicesToRegisterBasedOnTaskLabels(t *testing.T) {
	os.Setenv("MESOS_EXECUTOR_ID", "executor_id")
	os.Setenv("MESOS_FRAMEWORK_ID", "framework_id")
	os.Setenv("HOST", "hostname")
	defer os.Unsetenv("MESOS_EXECUTOR_ID")
	defer os.Unsetenv("MESOS_FRAMEWORK_ID")
	defer os.Unsetenv("HOST")

	s := state{Frameworks: []framework{framework{
		ID: "framework_id",
		Executors: []executor{executor{
			ID: "executor_id",
			Tasks: []task{task{
				Labels:    []label{label{Key: "consul", Value: "name"}},
				Discovery: discovery{Ports: ports{Ports: []port{port{Number: 1234}}}},
			}},
		}},
	}}}

	agentClient := &mockAgentClient{}
	agentClient.On("state").Return(s, nil)

	serviceProvider := ServiceProvider{
		agentClient: agentClient,
	}

	serviceInstances, err := serviceProvider.Get(context.Background())

	require.NoError(t, err)
	require.NotEmpty(t, serviceInstances)
	assert.Equal(t, "hostname_1234", serviceInstances[0].ID)
	assert.Equal(t, "name", serviceInstances[0].Name)
	assert.Equal(t, "hostname", serviceInstances[0].Host)
	assert.Equal(t, 1234, serviceInstances[0].Port)
}

func TestIfReturnsServicesToRegisterBasedOnPortLabels(t *testing.T) {
	os.Setenv("MESOS_EXECUTOR_ID", "executor_id")
	os.Setenv("MESOS_FRAMEWORK_ID", "framework_id")
	os.Setenv("HOST", "hostname")
	defer os.Unsetenv("MESOS_EXECUTOR_ID")
	defer os.Unsetenv("MESOS_FRAMEWORK_ID")
	defer os.Unsetenv("HOST")

	s := state{Frameworks: []framework{framework{
		ID: "framework_id",
		Executors: []executor{executor{
			ID: "executor_id",
			Tasks: []task{task{
				Labels: []label{label{Key: "consul", Value: "invalid-name"}},
				Discovery: discovery{Ports: ports{Ports: []port{port{
					Number: 1234,
					Labels: []label{label{Key: "consul", Value: "valid-name"}},
				}}}},
			}},
		}},
	}}}

	agentClient := &mockAgentClient{}
	agentClient.On("state").Return(s, nil)

	serviceProvider := ServiceProvider{
		agentClient: agentClient,
	}

	serviceInstances, err := serviceProvider.Get(context.Background())

	require.NoError(t, err)
	require.NotEmpty(t, serviceInstances)
	assert.Equal(t, "hostname_1234", serviceInstances[0].ID)
	assert.Equal(t, "valid-name", serviceInstances[0].Name)
	assert.Equal(t, "hostname", serviceInstances[0].Host)
	assert.Equal(t, 1234, serviceInstances[0].Port)
}

func TestIfNotPanicsWithEmptyPorts(t *testing.T) {
	os.Setenv("MESOS_EXECUTOR_ID", "executor_id")
	os.Setenv("MESOS_FRAMEWORK_ID", "framework_id")
	os.Setenv("HOST", "hostname")
	defer os.Unsetenv("MESOS_EXECUTOR_ID")
	defer os.Unsetenv("MESOS_FRAMEWORK_ID")
	defer os.Unsetenv("HOST")

	require.NotPanics(t, func() {
		s := state{Frameworks: []framework{framework{
			ID: "framework_id",
			Executors: []executor{executor{
				ID: "executor_id",
				Tasks: []task{task{
					Labels: []label{label{Key: "consul", Value: "name"}},
				}},
			}},
		}}}

		agentClient := &mockAgentClient{}
		agentClient.On("state").Return(s, nil)

		serviceProvider := ServiceProvider{
			agentClient: agentClient,
		}

		serviceInstances, err := serviceProvider.Get(context.Background())

		require.NoError(t, err)
		require.Empty(t, serviceInstances)
	})
}

func TestIfNotPanicsWithEmptyTasks(t *testing.T) {
	os.Setenv("MESOS_EXECUTOR_ID", "executor_id")
	os.Setenv("MESOS_FRAMEWORK_ID", "framework_id")
	os.Setenv("HOST", "hostname")
	defer os.Unsetenv("MESOS_EXECUTOR_ID")
	defer os.Unsetenv("MESOS_FRAMEWORK_ID")
	defer os.Unsetenv("HOST")

	require.NotPanics(t, func() {
		s := state{Frameworks: []framework{framework{
			ID: "framework_id",
			Executors: []executor{executor{
				ID: "executor_id",
			}},
		}}}

		agentClient := &mockAgentClient{}
		agentClient.On("state").Return(s, nil)

		serviceProvider := ServiceProvider{
			agentClient: agentClient,
		}

		serviceInstances, err := serviceProvider.Get(context.Background())

		require.Error(t, err)
		require.Empty(t, serviceInstances)
	})
}

func TestIfConvertsMesosLabelsToConsulTags(t *testing.T) {
	os.Setenv("MESOS_EXECUTOR_ID", "executor_id")
	os.Setenv("MESOS_FRAMEWORK_ID", "framework_id")
	os.Setenv("HOST", "hostname")
	defer os.Unsetenv("MESOS_EXECUTOR_ID")
	defer os.Unsetenv("MESOS_FRAMEWORK_ID")
	defer os.Unsetenv("HOST")

	s := state{Frameworks: []framework{framework{
		ID: "framework_id",
		Executors: []executor{executor{
			ID: "executor_id",
			Tasks: []task{task{
				Labels: []label{
					label{Key: "consul", Value: "name"},
					label{Key: "tag1", Value: "tag"},
					label{Key: "tag2", Value: "tag"},
				},
				Discovery: discovery{Ports: ports{Ports: []port{port{Number: 1234}}}},
			}},
		}},
	}}}

	agentClient := &mockAgentClient{}
	agentClient.On("state").Return(s, nil)

	serviceProvider := ServiceProvider{
		agentClient: agentClient,
	}

	serviceInstances, err := serviceProvider.Get(context.Background())

	require.NoError(t, err)
	require.NotEmpty(t, serviceInstances)
	assert.Equal(t, []string{"tag1", "tag2"}, serviceInstances[0].Tags)
}

type mockAgentClient struct {
	mock.Mock
}

func (ac *mockAgentClient) state() (state, error) {
	args := ac.Called()
	return args.Get(0).(state), args.Error(1)
}
