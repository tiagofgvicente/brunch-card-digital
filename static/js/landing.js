const LandingPage = {
    setup() {
        const { ref, computed, onMounted } = Vue;

        const theme = ref('light'); 
        const lang = ref('pt'); // Default Português
        const currentModal = ref(null); 
        const loading = ref(false);
        const errorMsg = ref('');
        const successMsg = ref(''); 
        
        // 👇 NOVA VARIÁVEL: Guarda o link de redirecionamento 👇
        const pendingRedirect = ref('');
        
        const storeTab = ref('login'); 
        const walletTab = ref('login'); 

        const storeConfig = ref({ name: 'Volto', logo: '' });
        const stats = ref({ total_cards: 0, total_stamps: 0, total_redeems: 0 });
        
        const loginForm = ref({ email: '', password: '' });
        const registerForm = ref({ name: '', email: '', password: '' });
        const walletForm = ref({ email: '', password: '' });
        const globalUserForm = ref({ firstName: '', lastName: '', email: '', phone: '', password: '', rgpd: false, marketing: false });

        const isWalletFormValid = computed(() => {
            const f = globalUserForm.value;
            return f.firstName.trim() !== '' &&
                   f.lastName.trim() !== '' &&
                   f.email.trim() !== '' &&
                   f.password.trim() !== '' &&
                   f.rgpd === true;
        });

        const translations = {
            en: {
                nav_wallet: "My Wallet", 
                nav_business: "For Businesses",
                tab_login: "Login",
                tab_register_store: "Start Free Trial",
                tab_register_wallet: "Create Account",
                btn_login: "Login",
                btn_enter: "Enter Dashboard",
                hero_title: "Turn customers <br> into regulars.", hero_subtitle: "The simplest loyalty platform for modern businesses.",
                stats_users: "Happy Users", stats_stamps: "Stamps Given", stats_rewards: "Rewards Claimed",
                modal_login_title: "Store Dashboard", password: "Password",
                modal_register_title: "Grow Your Business", modal_register_sub: "30 days free. Setup in 1 minute.",
                business_name: "Business Name", btn_launch: "Launch Platform",
                wallet_title: "My Wallet", wallet_sub: "Access all your cards securely."
            },
            pt: {
                nav_wallet: "Minha Carteira", 
                nav_business: "Para Negócios",
                tab_login: "Entrar",
                tab_register_store: "Teste Grátis",
                tab_register_wallet: "Criar Conta",
                btn_login: "Entrar",
                btn_enter: "Entrar no Painel",
                hero_title: "Transforme clientes <br> em fãs.", hero_subtitle: "A plataforma de fidelização mais simples para o seu negócio.",
                stats_users: "Clientes Felizes", stats_stamps: "Selos Dados", stats_rewards: "Prémios Dados",
                modal_login_title: "Gestão de Loja", password: "Palavra-passe",
                modal_register_title: "Comece a Crescer", modal_register_sub: "30 dias grátis. Configuração em 1 minuto.",
                business_name: "Nome do Negócio", btn_launch: "Lançar Plataforma",
                wallet_title: "Minha Carteira", wallet_sub: "Aceda aos seus cartões com segurança."
            }
        };

        const t = (key) => translations[lang.value][key] || key;

        const toggleTheme = () => {
            theme.value = theme.value === 'dark' ? 'light' : 'dark';
            document.documentElement.setAttribute('data-theme', theme.value);
            localStorage.setItem('theme', theme.value);
        };

        const toggleLang = () => {
            lang.value = lang.value === 'pt' ? 'en' : 'pt';
            localStorage.setItem('lang', lang.value);
        }

        const initSettings = () => {
            theme.value = localStorage.getItem('theme') || 'light';
            document.documentElement.setAttribute('data-theme', theme.value);
            lang.value = localStorage.getItem('lang') || 'pt';
        };

        const fetchConfig = async () => { try { const res = await fetch('/api/v1/system/config'); if(res.ok) storeConfig.value = await res.json(); } catch(e){} };
        const fetchStats = async () => { try { const res = await fetch('/api/v1/public/stats'); if(res.ok) stats.value = await res.json(); } catch(e){} };

        const handleLogin = async () => {
            loading.value = true; errorMsg.value = ''; successMsg.value = '';
            try {
                const res = await fetch('/api/v1/auth/login', { 
                    method: 'POST', 
                    headers: { 'Content-Type': 'application/json' }, 
                    body: JSON.stringify({ identifier: loginForm.value.email, password: loginForm.value.password }) 
                });
                
                if (res.ok) { 
                    const data = await res.json(); 
                    window.location.href = data.redirect; 
                } else { 
                    const errorText = await res.text();
                    if (errorText.includes("SUSPENDED")) {
                        errorMsg.value = lang.value === 'pt' ? "⛔ A sua conta encontra-se suspensa." : "⛔ Account suspended.";
                    } else if (errorText.includes("unverified")) {
                        errorMsg.value = lang.value === 'pt' ? "📧 Vá ao seu email para ativar a conta." : "📧 Please check your email to activate.";
                    } else {
                        errorMsg.value = lang.value === 'pt' ? "Email ou Password incorretos." : "Invalid email or password."; 
                    }
                }
            } catch (e) { errorMsg.value = lang.value === 'pt' ? "Erro de conexão." : "Connection error."; } 
            finally { loading.value = false; }
        };
        
        const handleRegister = async () => {
            loading.value = true; errorMsg.value = ''; successMsg.value = '';
            try {
                const res = await fetch('/api/v1/public/register', { 
                    method: 'POST', headers: {'Content-Type': 'application/json'}, 
                    body: JSON.stringify(registerForm.value) 
                });

                if (res.ok) { 
                    const data = await res.json(); 
                    if (data.status === "pending_verification") {
                        successMsg.value = data.message;
                        pendingRedirect.value = data.redirect; // 👈 Guarda o redirecionamento
                        registerForm.value = { name: '', email: '', password: '' };
                    } else {
                        window.location.href = data.redirect; 
                    }
                } else { 
                    const txt = await res.text(); 
                    errorMsg.value = txt.trim() || (lang.value === 'pt' ? "Erro ao criar conta." : "Error creating account."); 
                }
            } catch (e) { errorMsg.value = lang.value === 'pt' ? "Erro de conexão." : "Connection error."; } 
            finally { loading.value = false; }
        };

        const handleWalletRegister = async () => {
            loading.value = true; errorMsg.value = ''; successMsg.value = '';
            try {
                const res = await fetch('/api/v1/public/wallet-register', {
                    method: 'POST', headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(globalUserForm.value)
                });

                if (res.ok) {
                    const data = await res.json();
                    if (data.status === "pending_verification") {
                        successMsg.value = data.message;
                        pendingRedirect.value = data.redirect; // 👈 Guarda o redirecionamento
                        globalUserForm.value = { firstName: '', lastName: '', email: '', phone: '', password: '', rgpd: false, marketing: false };
                    } else {
                        window.location.href = data.redirect;
                    }
                } else {
                    const txt = await res.text();
                    errorMsg.value = txt.trim() || (lang.value === 'pt' ? "Erro ao criar conta." : "Error creating account.");
                }
            } catch (e) { errorMsg.value = lang.value === 'pt' ? "Erro de ligação." : "Connection error."; } 
            finally { loading.value = false; }
        };

        const handleWalletLogin = async () => {
            loading.value = true; errorMsg.value = ''; successMsg.value = '';
            try {
                const res = await fetch('/api/v1/public/wallet-login', {
                    method: 'POST', headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ Email: walletForm.value.email, Password: walletForm.value.password })
                });

                if (res.ok) {
                    const data = await res.json();
                    window.location.href = data.redirect;
                } else {
                    const txt = await res.text();
                    if (txt.includes("unverified")) {
                        errorMsg.value = lang.value === 'pt' ? "📧 Vá ao seu email para ativar a conta." : "📧 Please check your email to activate.";
                    } else {
                        errorMsg.value = lang.value === 'pt' ? "Email ou Password incorretos." : "Invalid credentials.";
                    }
                }
            } catch (e) { errorMsg.value = lang.value === 'pt' ? "Erro de ligação." : "Connection error."; } 
            finally { loading.value = false; }
        };

        // 👇 NOVA FUNÇÃO: Segue para a App em vez de fechar o modal
        const proceedToApp = () => {
            if (pendingRedirect.value) {
                window.location.href = pendingRedirect.value;
            } else {
                currentModal.value = null;
                successMsg.value = '';
            }
        };

        onMounted(() => { initSettings(); fetchConfig(); fetchStats(); });

        return {
            theme, lang, currentModal, loading, errorMsg, successMsg, storeConfig, stats,
            loginForm, registerForm, walletForm, walletTab, storeTab, globalUserForm,
            isWalletFormValid, t, toggleTheme, toggleLang, 
            handleLogin, handleRegister, handleWalletRegister, handleWalletLogin,
            proceedToApp, // 👈 Função exportada
            
            openLogin: () => { currentModal.value = 'store'; storeTab.value = 'login'; errorMsg.value = ''; successMsg.value = ''; },
            openRegister: () => { currentModal.value = 'store'; storeTab.value = 'register'; errorMsg.value = ''; successMsg.value = ''; },
            
            openWallet: () => { 
                currentModal.value = 'wallet'; 
                walletTab.value = 'login'; 
                walletForm.value = { email: '', password: '' };
                globalUserForm.value = { firstName: '', lastName: '', email: '', phone: '', password: '', rgpd: false, marketing: false };
                errorMsg.value = ''; successMsg.value = '';
            },
            closeModal: () => { currentModal.value = null; successMsg.value = ''; errorMsg.value = ''; pendingRedirect.value = ''; }
        };
    }
};

Vue.createApp(LandingPage).mount('#app');