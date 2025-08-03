import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Layout, Menu, Spin, Typography, Alert } from 'antd';
import ScenarioViewer from './components/ScenarioViewer';

const { Header, Content, Sider } = Layout;
const { Title } = Typography;

// Configure axios to proxy requests to the backend container
const apiClient = axios.create({
    baseURL: 'http://localhost:8080/api', // Assumes local dev environment
});

function App() {
    const [scenarios, setScenarios] = useState([]);
    const [selectedScenarioId, setSelectedScenarioId] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    // Fetch all available scenarios on component mount
    useEffect(() => {
        apiClient.get('/scenarios')
            .then(response => {
                setScenarios(response.data);
                setLoading(false);
            })
            .catch(err => {
                console.error("Failed to fetch scenarios:", err);
                setError("Could not connect to the backend. Please ensure it's running.");
                setLoading(false);
            });
    }, []);

    const handleMenuClick = ({ key }) => {
        setSelectedScenarioId(key);
    };

    return (
        <Layout style={{ minHeight: '100vh' }}>
            <Header style={{ color: 'white' }}>
                <Title level={3} style={{ color: 'inherit', margin: '16px 0' }}>Backend Problem Playground</Title>
            </Header>
            <Layout>
                <Sider width={300} style={{ background: '#fff' }}>
                    {loading ? (
                        <div style={{ padding: '24px', textAlign: 'center' }}><Spin /></div>
                    ) : (
                        <Menu
                            mode="inline"
                            style={{ height: '100%', borderRight: 0 }}
                            onClick={handleMenuClick}
                            items={scenarios.map(s => ({ key: s.id, label: s.title }))}
                            selectedKeys={[selectedScenarioId]}
                        />
                    )}
                </Sider>
                <Layout style={{ padding: '0 24px 24px' }}>
                    <Content
                        style={{
                            padding: 24,
                            margin: 0,
                            minHeight: 280,
                            background: '#fff',
                        }}
                    >
                        {error && <Alert message="Error" description={error} type="error" showIcon />}
                        {!selectedScenarioId && !error && (
                            <div style={{ textAlign: 'center', paddingTop: '50px' }}>
                                <Title level={4}>Welcome!</Title>
                                <p>Please select a scenario from the left menu to begin.</p>
                            </div>
                        )}
                        {selectedScenarioId && <ScenarioViewer scenarioId={selectedScenarioId} />}
                    </Content>
                </Layout>
            </Layout>
        </Layout>
    );
}

export default App;