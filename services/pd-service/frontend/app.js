
const API_BASE = '/api';

// State
let currentOrg = '';
let services = [];
let escalationPolicies = [];
let githubRepos = [];
let slackUsers = [];

// Initialize app
document.addEventListener('DOMContentLoaded', () => {
    initNavigation();
    loadOrganizations();
    setupEventListeners();
});

// Navigation
function initNavigation() {
    const navItems = document.querySelectorAll('.nav-item');
    navItems.forEach(item => {
        item.addEventListener('click', () => {
            const panel = item.dataset.panel;
            switchPanel(panel);
        });
    });
}

function switchPanel(panelName) {
    // Update nav
    document.querySelectorAll('.nav-item').forEach(item => {
        item.classList.remove('active');
    });
    document.querySelector(`[data-panel="${panelName}"]`).classList.add('active');
    
    // Update panels
    document.querySelectorAll('.panel').forEach(panel => {
        panel.classList.remove('active');
    });
    document.getElementById(`${panelName}-panel`).classList.add('active');
    
    // Update title
    const titles = {
        home: 'Home',
        scorecard: 'Scorecard',
        trigger: 'Trigger Incident'
    };
    document.getElementById('panel-title').textContent = titles[panelName];
    
    // Load panel data
    if (panelName === 'scorecard') {
        loadMetrics();
    } else if (panelName === 'trigger') {
        loadServicesForTrigger();
    }
}

// Event Listeners
function setupEventListeners() {
    // Org selector
    document.getElementById('org-select').addEventListener('change', (e) => {
        currentOrg = e.target.value;
        loadServices();
    });

    // Add service button
    document.getElementById('add-service-btn').addEventListener('click', openAddServiceModal);

    // Modal close
    document.querySelector('.close').addEventListener('click', closeAddServiceModal);
    document.getElementById('cancel-add-service').addEventListener('click', closeAddServiceModal);

    // Add service form
    const addServiceForm = document.getElementById('add-service-form');
    if (addServiceForm) {
        console.log('Add service form found, attaching event listener');
        addServiceForm.addEventListener('submit', (e) => {
            console.log('Form submit event triggered');
            handleAddService(e);
        });
    } else {
        console.error('Add service form not found!');
    }

    // Trigger form
    document.getElementById('trigger-form').addEventListener('submit', handleTriggerIncident);

    // Refresh metrics
    document.getElementById('refresh-metrics-btn').addEventListener('click', loadMetrics);
}

// API Calls
async function apiCall(endpoint, options = {}) {
    try {
        console.log('API Call:', endpoint, options);
        const response = await fetch(`${API_BASE}${endpoint}`, options);
        console.log('API Response:', response.status, response.statusText);

        if (!response.ok) {
            let errorMessage = 'API request failed';
            try {
                const error = await response.json();
                errorMessage = error.error || errorMessage;
            } catch (e) {
                errorMessage = `${response.status} ${response.statusText}`;
            }
            throw new Error(errorMessage);
        }

        const data = await response.json();
        console.log('API Response Data:', data);
        return data;
    } catch (error) {
        console.error('API Error:', error);
        throw error;
    }
}

async function loadOrganizations() {
    try {
        const orgs = await apiCall('/organizations');
        const select = document.getElementById('org-select');
        select.innerHTML = orgs.map(org => 
            `<option value="${org.name}">${org.name}</option>`
        ).join('');
        currentOrg = orgs[0]?.name || '';
        loadServices();
    } catch (error) {
        console.error('Failed to load organizations:', error);
    }
}

async function loadServices() {
    try {
        const container = document.getElementById('services-list');
        container.innerHTML = '<div class="loading">Loading services...</div>';
        
        services = await apiCall(`/services?org=${currentOrg}`);
        
        if (services.length === 0) {
            container.innerHTML = '<div class="loading">No services added yet. Click "Add Service" to get started.</div>';
            return;
        }
        
        container.innerHTML = services.map(service => `
            <div class="service-card">
                <div class="service-card-header">
                    <div>
                        <h4>${service.name}</h4>
                        <div class="service-card-info">📦 ${service.github_repo}</div>
                        <div class="service-card-info">👤 ${service.slack_assignee}</div>
                    </div>
                </div>
                <div class="service-card-actions">
                    <button class="btn btn-small btn-secondary" onclick="viewServiceMetrics('${service.id}')">
                        View Metrics
                    </button>
                    <button class="btn btn-small delete-btn" onclick="deleteService('${service.id}')">
                        Delete
                    </button>
                </div>
            </div>
        `).join('');
    } catch (error) {
        document.getElementById('services-list').innerHTML = 
            `<div class="loading">Error loading services: ${error.message}</div>`;
    }
}

async function deleteService(id) {
    if (!confirm('Are you sure you want to delete this service?')) {
        return;
    }
    
    try {
        await apiCall(`/services/${id}`, { method: 'DELETE' });
        loadServices();
    } catch (error) {
        alert('Failed to delete service: ' + error.message);
    }
}

function viewServiceMetrics(id) {
    switchPanel('scorecard');
    // Scroll to the service metrics
    setTimeout(() => {
        const element = document.getElementById(`metrics-${id}`);
        if (element) {
            element.scrollIntoView({ behavior: 'smooth' });
        }
    }, 100);
}

// Add Service Modal
async function openAddServiceModal() {
    const modal = document.getElementById('add-service-modal');
    modal.classList.add('active');

    // Load GitHub repos
    try {
        githubRepos = await apiCall(`/github/repos?org=${currentOrg}`);
        const select = document.getElementById('github-repo');
        select.innerHTML = '<option value="">Select GitHub repository...</option>' +
            githubRepos.map(r => `<option value="${r.full_name}">${r.name}</option>`).join('');
    } catch (error) {
        console.error('Failed to load GitHub repos:', error);
    }

    // Load Slack users
    try {
        slackUsers = await apiCall('/slack/users');
        const select = document.getElementById('slack-assignee');
        select.innerHTML = '<option value="">Select Slack user...</option>' +
            slackUsers.map(u => `<option value="${u.id}">${u.real_name || u.name} (${u.email})</option>`).join('');
    } catch (error) {
        console.error('Failed to load Slack users:', error);
    }
}

function closeAddServiceModal() {
    document.getElementById('add-service-modal').classList.remove('active');
    document.getElementById('add-service-form').reset();
}

async function handleAddService(e) {
    e.preventDefault();

    console.log('Add service form submitted');

    const submitBtn = document.getElementById('submit-add-service');
    const originalText = submitBtn.textContent;
    submitBtn.disabled = true;
    submitBtn.textContent = 'Adding...';

    const name = document.getElementById('service-name').value;
    const githubRepo = document.getElementById('github-repo').value;
    const slackAssigneeId = document.getElementById('slack-assignee').value;

    console.log('Form values:', { name, githubRepo, slackAssigneeId });

    // Validation
    if (!name || !githubRepo || !slackAssigneeId) {
        alert('Please fill in all required fields');
        submitBtn.disabled = false;
        submitBtn.textContent = originalText;
        return;
    }

    const slackUser = slackUsers.find(u => u.id === slackAssigneeId);

    const service = {
        name,
        github_repo: githubRepo,
        slack_assignee: slackUser?.real_name || slackUser?.name || '',
        slack_assignee_id: slackAssigneeId,
        org_name: currentOrg
    };

    console.log('Creating service:', service);
    console.log('Service JSON:', JSON.stringify(service));

    try {
        console.log('About to call API...');
        const result = await apiCall('/services', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(service)
        });
        console.log('API call completed');

        console.log('Service created successfully:', result);

        closeAddServiceModal();
        loadServices();

        // Show success message
        alert('✅ Service added successfully!');
    } catch (error) {
        console.error('Failed to add service:', error);
        alert('❌ Failed to add service: ' + error.message);
    } finally {
        submitBtn.disabled = false;
        submitBtn.textContent = originalText;
    }
}

// Metrics
async function loadMetrics() {
    try {
        const container = document.getElementById('metrics-overview');
        container.innerHTML = '<div class="loading">Loading metrics...</div>';

        const metrics = await apiCall(`/metrics?org=${currentOrg}`);

        if (metrics.length === 0) {
            container.innerHTML = '<div class="loading">No services to show metrics for.</div>';
            document.getElementById('metrics-details').innerHTML = '';
            return;
        }

        // Calculate overview
        const totalServices = metrics.length;
        const totalIncidents = metrics.reduce((sum, m) => sum + m.total_incidents, 0);
        const totalOpen = metrics.reduce((sum, m) => sum + m.open_incidents, 0);
        const avgResolveTime = metrics.reduce((sum, m) => sum + m.avg_time_to_resolve, 0) / metrics.length;

        container.innerHTML = `
            <div class="metric-card">
                <h4>Total Services</h4>
                <div class="metric-value">${totalServices}</div>
            </div>
            <div class="metric-card">
                <h4>Total Incidents</h4>
                <div class="metric-value">${totalIncidents}</div>
            </div>
            <div class="metric-card">
                <h4>Open Incidents</h4>
                <div class="metric-value">${totalOpen}</div>
            </div>
            <div class="metric-card">
                <h4>Avg Resolve Time</h4>
                <div class="metric-value">${avgResolveTime.toFixed(1)}<span style="font-size: 1rem;">min</span></div>
            </div>
        `;

        // Render individual service metrics
        renderServiceMetrics(metrics);
    } catch (error) {
        document.getElementById('metrics-overview').innerHTML =
            `<div class="loading">Error loading metrics: ${error.message}</div>`;
    }
}

function renderServiceMetrics(metrics) {
    const container = document.getElementById('metrics-details');

    container.innerHTML = metrics.map(m => `
        <div class="service-metrics-card" id="metrics-${m.service_id}">
            <div class="service-metrics-header">
                <div>
                    <h4>${m.service_name}</h4>
                    <div style="color: var(--text-secondary); font-size: 0.9rem;">
                        👤 ${m.assignee_name}
                    </div>
                </div>
                <div>
                    ${m.open_incidents > 0 ?
                        `<span class="badge badge-danger">${m.open_incidents} Open</span>` :
                        `<span class="badge badge-success">No Open Incidents</span>`
                    }
                </div>
            </div>

            <div class="metrics-grid">
                <div class="metric-item">
                    <div class="metric-item-label">Total Incidents</div>
                    <div class="metric-item-value">${m.total_incidents}</div>
                </div>
                <div class="metric-item">
                    <div class="metric-item-label">High Priority</div>
                    <div class="metric-item-value">${m.high_priority}</div>
                </div>
                <div class="metric-item">
                    <div class="metric-item-label">Avg Resolve Time</div>
                    <div class="metric-item-value">${m.avg_time_to_resolve.toFixed(1)} min</div>
                </div>
                <div class="metric-item">
                    <div class="metric-item-label">Avg Response Time</div>
                    <div class="metric-item-value">${m.avg_time_to_respond.toFixed(1)} min</div>
                </div>
            </div>

            <div class="chart-container">
                <canvas id="chart-${m.service_id}"></canvas>
            </div>
        </div>
    `).join('');

    // Render charts
    metrics.forEach(m => {
        renderChart(m);
    });
}

function renderChart(metrics) {
    const ctx = document.getElementById(`chart-${metrics.service_id}`);
    if (!ctx) return;

    new Chart(ctx, {
        type: 'bar',
        data: {
            labels: ['Total', 'Open', 'High Priority'],
            datasets: [{
                label: 'Incidents',
                data: [metrics.total_incidents, metrics.open_incidents, metrics.high_priority],
                backgroundColor: [
                    'rgba(6, 193, 103, 0.5)',
                    'rgba(220, 53, 69, 0.5)',
                    'rgba(255, 193, 7, 0.5)'
                ],
                borderColor: [
                    'rgb(6, 193, 103)',
                    'rgb(220, 53, 69)',
                    'rgb(255, 193, 7)'
                ],
                borderWidth: 1
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: {
                        color: '#a0a4b8'
                    },
                    grid: {
                        color: '#3a3f4b'
                    }
                },
                x: {
                    ticks: {
                        color: '#a0a4b8'
                    },
                    grid: {
                        color: '#3a3f4b'
                    }
                }
            }
        }
    });
}

// Trigger Incident
async function loadServicesForTrigger() {
    try {
        const select = document.getElementById('trigger-service');
        select.innerHTML = '<option value="">Loading...</option>';

        const svcs = await apiCall(`/services?org=${currentOrg}`);

        if (svcs.length === 0) {
            select.innerHTML = '<option value="">No services available</option>';
            return;
        }

        select.innerHTML = '<option value="">Select a service...</option>' +
            svcs.map(s => `<option value="${s.id}">${s.name}</option>`).join('');
    } catch (error) {
        console.error('Failed to load services:', error);
    }
}

async function handleTriggerIncident(e) {
    e.preventDefault();

    const serviceId = document.getElementById('trigger-service').value;
    const title = document.getElementById('trigger-title').value;
    const description = document.getElementById('trigger-description').value;
    const priority = document.getElementById('trigger-priority').value;

    const resultDiv = document.getElementById('trigger-result');
    resultDiv.innerHTML = '<div class="loading">Creating incident...</div>';

    try {
        const result = await apiCall('/incidents/trigger', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                service_id: serviceId,
                title,
                description,
                priority
            })
        });

        resultDiv.className = 'trigger-result success';
        resultDiv.innerHTML = `
            <h4>✅ Incident Created Successfully!</h4>
            <p><strong>Incident ID:</strong> ${result.incident_id}</p>
            <p><strong>URL:</strong> <a href="${result.incident_url}" target="_blank" style="color: var(--primary-color);">${result.incident_url}</a></p>
            <p>${result.message}</p>
        `;

        // Reset form
        document.getElementById('trigger-form').reset();
    } catch (error) {
        resultDiv.className = 'trigger-result error';
        resultDiv.innerHTML = `
            <h4>❌ Failed to Create Incident</h4>
            <p>${error.message}</p>
        `;
    }
}

