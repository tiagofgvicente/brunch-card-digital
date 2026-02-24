const LandingPage = {
    setup() {
        const { ref, computed, onMounted } = Vue;

        const theme = ref('light'); 
        const lang = ref('pt'); 
        const currentModal = ref(null); 
        const loading = ref(false);
        const errorMsg = ref('');
        const successMsg = ref(''); 
        
        // Guarda o link de redirecionamento para o botão "Entendido, Continuar"
        const pendingRedirect = ref('');
        
        const storeTab = ref('login'); 
        const walletTab = ref('login'); 

        // --- ESTADOS PARA FORGOT PASSWORD ---
        const forgotMode = ref(false);
        const forgotEmail = ref('');
        const resetToken = ref('');
        const resetType = ref(''); // 'store' ou 'wallet'
        const resetPasswordForm = ref({ password: '' });

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
                nav_wallet: "My Wallet", nav_business: "For Businesses",
                tab_login: "Login", tab_register_store: "Start Free Trial", tab_register_wallet: "Create Account",
                btn_login: "Login", btn_enter: "Enter Dashboard",
                hero_title: "Turn customers <br> into regulars.", hero_subtitle: "The simplest loyalty platform for modern businesses.",
                stats_users: "Happy Users", stats_stamps: "Stamps Given", stats_rewards: "Rewards Claimed",
                modal_login_title: "Store Dashboard", password: "Password",
                modal_register_title: "Grow Your Business", modal_register_sub: "30 days free. Setup in 1 minute.",
                business_name: "Business Name", btn_launch: "Launch Platform",
                wallet_title: "My Wallet", wallet_sub: "Access all your cards securely.",
                forgot_pass: "Forgot your password?", forgot_title: "Recover Password", forgot_sub: "Enter your email. We'll send you a recovery link.",
                btn_send_link: "Send Link", back_login: "Back to Login", reset_title: "New Password", reset_sub: "Choose a secure new password.", btn_save_pass: "Save Password"
            },
            pt: {
                nav_wallet: "Minha Carteira", nav_business: "Para Negócios",
                tab_login: "Entrar", tab_register_store: "Teste Grátis", tab_register_wallet: "Criar Conta",
                btn_login: "Entrar", btn_enter: "Entrar no Painel",
                hero_title: "Transforme clientes <br> em fãs.", hero_subtitle: "A plataforma de fidelização mais simples para o seu negócio.",
                stats_users: "Clientes Felizes", stats_stamps: "Selos Dados", stats_rewards: "Prémios Dados",
                modal_login_title: "Gestão de Loja", password: "Palavra-passe",
                modal_register_title: "Comece a Crescer", modal_register_sub: "30 dias grátis. Configuração em 1 minuto.",
                business_name: "Nome do Negócio", btn_launch: "Lançar Plataforma",
                wallet_title: "Minha Carteira", wallet_sub: "Aceda aos seus cartões com segurança.",
                forgot_pass: "Esqueceu-se da palavra-passe?", forgot_title: "Recuperar Palavra-passe", forgot_sub: "Insira o seu email. Vamos enviar-lhe um link de recuperação.",
                btn_send_link: "Enviar Link", back_login: "Voltar ao Login", reset_title: "Nova Palavra-passe", reset_sub: "Escolha uma nova palavra-passe segura.", btn_save_pass: "Guardar Palavra-passe"
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
                    } else {
                        errorMsg.value = lang.value === 'pt' ? "Email ou Password incorretos." : "Invalid email or password."; 
                    }
                }
            } catch (e) { errorMsg.value = lang.value === 'pt' ? "Erro de conexão." : "Connection error."; } 
            finally { loading.value = false; }
        };
        
        const handleRegister = async () => {
            loading.value = true; errorMsg.value = ''; successMsg.value = '';
            const emailToLogin = registerForm.value.email;
            const passToLogin = registerForm.value.password;

            try {
                const res = await fetch('/api/v1/public/register', { 
                    method: 'POST', headers: {'Content-Type': 'application/json'}, 
                    body: JSON.stringify(registerForm.value) 
                });

                if (res.ok) { 
                    const data = await res.json(); 
                    if (data.status === "pending_verification") {
                        successMsg.value = data.message;
                        
                        try {
                            const loginRes = await fetch('/api/v1/auth/login', { 
                                method: 'POST', headers: { 'Content-Type': 'application/json' }, 
                                body: JSON.stringify({ identifier: emailToLogin, password: passToLogin }) 
                            });
                            if(loginRes.ok) {
                                const loginData = await loginRes.json();
                                pendingRedirect.value = loginData.redirect;
                            } else {
                                pendingRedirect.value = data.redirect || '/'; 
                            }
                        } catch(e) { pendingRedirect.value = data.redirect || '/'; }

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
            const emailToLogin = globalUserForm.value.email;
            const passToLogin = globalUserForm.value.password;

            try {
                const res = await fetch('/api/v1/public/wallet-register', {
                    method: 'POST', headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(globalUserForm.value)
                });

                if (res.ok) {
                    const data = await res.json();
                    if (data.status === "pending_verification") {
                        successMsg.value = data.message;
                        
                        try {
                            const loginRes = await fetch('/api/v1/public/wallet-login', {
                                method: 'POST', headers: { 'Content-Type': 'application/json' },
                                body: JSON.stringify({ Email: emailToLogin, Password: passToLogin })
                            });
                            if (loginRes.ok) {
                                const loginData = await loginRes.json();
                                pendingRedirect.value = loginData.redirect;
                            } else {
                                pendingRedirect.value = '/card'; 
                            }
                        } catch(e) { pendingRedirect.value = '/card'; }

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
                    errorMsg.value = lang.value === 'pt' ? "Email ou Password incorretos." : "Invalid credentials.";
                }
            } catch (e) { errorMsg.value = lang.value === 'pt' ? "Erro de ligação." : "Connection error."; } 
            finally { loading.value = false; }
        };

        // --- MÉTODOS DE FORGOT PASSWORD ---
        const handleForgotPassword = async () => {
            if (!forgotEmail.value) return;
            loading.value = true; errorMsg.value = ''; successMsg.value = '';
            const endpoint = currentModal.value === 'store' ? '/api/v1/auth/forgot-password' : '/api/v1/public/wallet-forgot-password';
            
            try {
                await fetch(endpoint, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ email: forgotEmail.value }) });
                successMsg.value = lang.value === 'pt' ? "Se o email existir, receberá um link em breve." : "If the email exists, you will receive a link shortly.";
            } catch (e) { errorMsg.value = lang.value === 'pt' ? "Erro de ligação." : "Connection error."; } 
            finally { loading.value = false; forgotEmail.value = ''; }
        };

        const handleResetPassword = async () => {
            loading.value = true; errorMsg.value = ''; successMsg.value = '';
            const endpoint = resetType.value === 'store' ? '/api/v1/auth/reset-password' : '/api/v1/public/wallet-reset-password';
            
            try {
                const res = await fetch(endpoint, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ token: resetToken.value, password: resetPasswordForm.value.password }) });
                if (res.ok) {
                    successMsg.value = lang.value === 'pt' ? "Palavra-passe alterada com sucesso! Já pode fazer login." : "Password reset successful! You can now login.";
                    resetToken.value = '';
                } else {
                    const txt = await res.text();
                    errorMsg.value = txt || (lang.value === 'pt' ? "Erro ao redefinir." : "Error resetting.");
                }
            } catch (e) { errorMsg.value = lang.value === 'pt' ? "Erro de ligação." : "Connection error."; } 
            finally { loading.value = false; }
        };

        const proceedToApp = () => {
            if (pendingRedirect.value) {
                window.location.href = pendingRedirect.value;
            } else if (resetToken.value === '') { 
                currentModal.value = null; 
                successMsg.value = ''; 
            } else { 
                successMsg.value = ''; 
            }
        };

        onMounted(() => { 
            initSettings(); fetchConfig(); fetchStats(); 
            
            // DETETA SE O UTILIZADOR VEM DO LINK DO EMAIL DE RECUPERAÇÃO
            const urlParams = new URLSearchParams(window.location.search);
            const tokenParam = urlParams.get('reset_token');
            const typeParam = urlParams.get('type');
            if (tokenParam && typeParam) {
                resetToken.value = tokenParam;
                resetType.value = typeParam;
                currentModal.value = 'reset_password';
            }
        });

        return {
            theme, lang, currentModal, loading, errorMsg, successMsg, storeConfig, stats,
            loginForm, registerForm, walletForm, walletTab, storeTab, globalUserForm,
            isWalletFormValid, t, toggleTheme, toggleLang, 
            handleLogin, handleRegister, handleWalletRegister, handleWalletLogin, proceedToApp,
            
            // Variáveis e funções exportadas para o template
            forgotMode, forgotEmail, resetToken, resetType, resetPasswordForm, handleForgotPassword, handleResetPassword,
            
            openLogin: () => { currentModal.value = 'store'; storeTab.value = 'login'; errorMsg.value = ''; successMsg.value = ''; forgotMode.value = false; },
            openRegister: () => { currentModal.value = 'store'; storeTab.value = 'register'; errorMsg.value = ''; successMsg.value = ''; forgotMode.value = false; },
            
            openWallet: () => { 
                currentModal.value = 'wallet'; 
                walletTab.value = 'login'; 
                walletForm.value = { email: '', password: '' };
                globalUserForm.value = { firstName: '', lastName: '', email: '', phone: '', password: '', rgpd: false, marketing: false };
                errorMsg.value = ''; successMsg.value = ''; forgotMode.value = false;
            },
            closeModal: () => { currentModal.value = null; successMsg.value = ''; errorMsg.value = ''; pendingRedirect.value = ''; forgotMode.value = false; resetToken.value = ''; }
        };
    }
};

Vue.createApp(LandingPage).mount('#app');