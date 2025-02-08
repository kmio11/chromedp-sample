const config = {
    authority: 'http://localhost:7080/realms/realm01/protocol/openid-connect',
    client_id: 'app',
    redirect_uri: 'http://localhost:3000/callback',
    response_type: 'code',
    scope: 'openid profile offline_access'
};

document.getElementById('loginButton').addEventListener('click', () => {
    // Generate random state and nonce
    const state = Math.random().toString(36).substring(7);
    const nonce = Math.random().toString(36).substring(7);
    
    // Store state in sessionStorage
    sessionStorage.setItem('auth_state', state);
    
    // Redirect to authorization endpoint
    const authUrl = new URL(config.authority + '/auth');
    authUrl.searchParams.append('client_id', config.client_id);
    authUrl.searchParams.append('redirect_uri', config.redirect_uri);
    authUrl.searchParams.append('response_type', config.response_type);
    authUrl.searchParams.append('scope', config.scope);
    authUrl.searchParams.append('state', state);
    authUrl.searchParams.append('nonce', nonce);
    
    window.location.href = authUrl.toString();
});

// Handle authorization code callback
if (window.location.pathname === '/callback') {
    const urlParams = new URLSearchParams(window.location.search);
    const code = urlParams.get('code');
    const state = urlParams.get('state');
    
    if (state === sessionStorage.getItem('auth_state')) {
        // Exchange code for tokens
        fetch(config.authority + '/token', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: new URLSearchParams({
                grant_type: 'authorization_code',
                code: code,
                redirect_uri: config.redirect_uri,
                client_id: config.client_id
            })
        })
        .then(response => response.json())
        .then(data => {
            // Store tokens in localStorage
            localStorage.setItem('access_token', data.access_token);
            localStorage.setItem('id_token', data.id_token);
            localStorage.setItem('refresh_token', data.refresh_token);
            // Redirect to protected page
            window.location.href = '/dashboard.html';
        });
    }
}
