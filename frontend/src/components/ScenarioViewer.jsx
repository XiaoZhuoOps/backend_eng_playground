import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Spin, Typography, Card, Button, Row, Col, Alert, Descriptions } from 'antd';

const { Title, Paragraph, Text } = Typography;

const apiClient = axios.create({
    baseURL: 'http://localhost:8080/api',
});

const ScenarioViewer = ({ scenarioId }) => {
    const [scenario, setScenario] = useState(null);
    const [state, setState] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [actionLoading, setActionLoading] = useState(false);

    // Fetch scenario details and initial state
    useEffect(() => {
        if (!scenarioId) return;
        setLoading(true);
        Promise.all([
            apiClient.get(`/scenarios/${scenarioId}`),
            apiClient.get(`/scenarios/${scenarioId}/state`)
        ]).then(([scenarioRes, stateRes]) => {
            setScenario(scenarioRes.data);
            setState(stateRes.data);
            setLoading(false);
        }).catch(err => {
            console.error("Failed to fetch scenario data:", err);
            setError("Could not load scenario data.");
            setLoading(false);
        });
    }, [scenarioId]);

    // Poll for state updates
    useEffect(() => {
        if (!scenarioId) return;
        const interval = setInterval(() => {
            apiClient.get(`/scenarios/${scenarioId}/state`)
                .then(res => setState(res.data))
                .catch(err => console.error("State poll failed:", err));
        }, 2000); // Poll every 2 seconds
        return () => clearInterval(interval);
    }, [scenarioId]);

    const handleActionClick = (actionId) => {
        setActionLoading(true);
        apiClient.post(`/scenarios/${scenarioId}/actions/${actionId}`)
            .then(() => {
                // State will update on the next poll
                setActionLoading(false);
            })
            .catch(err => {
                console.error(`Action ${actionId} failed:`, err);
                setActionLoading(false);
            });
    };

    if (!scenarioId) return null;
    if (loading) return <div style={{ textAlign: 'center', padding: '50px' }}><Spin size="large" /></div>;
    if (error) return <Alert message="Error" description={error} type="error" showIcon />;

    return (
        <div>
            <Title level={2}>{scenario.title}</Title>
            <Paragraph>{scenario.problem_description}</Paragraph>
            <Title level={4}>Solution</Title>
            <Paragraph>{scenario.solution_description}</Paragraph>

            <Row gutter={16}>
                <Col span={12}>
                    <Card title="Actions">
                        {scenario.actions.map(action => (
                            <Button
                                key={action.id}
                                onClick={() => handleActionClick(action.id)}
                                style={{ marginRight: '10px', marginBottom: '10px' }}
                                loading={actionLoading}
                            >
                                {action.name}
                            </Button>
                        ))}
                    </Card>
                </Col>
                <Col span={12}>
                    <Card title="Live Dashboard">
                        {state && Object.entries(state).map(([key, value]) => (
                            <Descriptions key={key} bordered column={1} size="small" style={{ marginBottom: 16 }}>
                                <Descriptions.Item label={key}>
                                    <pre style={{ margin: 0, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
                                        {typeof value === 'object' ? JSON.stringify(value, null, 2) : value}
                                    </pre>
                                </Descriptions.Item>
                            </Descriptions>
                        ))}
                    </Card>
                </Col>
            </Row>
        </div>
    );
};

export default ScenarioViewer;