const LandingPage = {
    setup() {
        const { ref, onMounted } = Vue;

        // --- ESTADO ---
        const theme = ref('light'); 
        const lang = ref('pt'); // Default Português
        const currentModal = ref(null); 
        const loading = ref(false);
        const errorMsg = ref('');
        const walletStep = ref(1); // 1 = Email, 2 = Código

        // --- DADOS ---
        const storeConfig = ref({ name: 'Volto', logo: '' });
        const stats = ref({ total_cards: 0, total_stamps: 0, total_redeems: 0 });
        
        // --- FORMULÁRIOS ---
        const loginForm = ref({ email: '', password: '' });
        const registerForm = ref({ name: '', email: '', password: '' });
        const walletForm = ref({ email: '', code: '' });

        // --- TRADUÇÕES ---
        const translations = {
            en: {
                wallet_btn: "My Wallet",
                login_btn: "Store Login",
                register_btn: "Start Free Trial",
                hero_title: "Turn customers <br> into regulars.",
                hero_subtitle: "The simplest loyalty platform for modern businesses.",
                stats_users: "Happy Users",
                stats_stamps: "Stamps Given",
                stats_rewards: "Rewards Claimed",
                modal_login_title: "Store Dashboard",
                password: "Password",
                btn_enter: "Enter Dashboard",
                new_business: "New business?",
                create_free_acc: "Create Free Account",
                modal_register_title: "Grow Your Business",
                modal_register_sub: "30 days free. Setup in 1 minute.",
                business_name: "Business Name",
                btn_launch: "Launch Platform 🚀",
                has_account: "Have an account?",
                btn_login: "Login",
                wallet_title: "My Wallet",
                wallet_sub: "Access all your cards securely.",
                btn_send_code: "Send Access Code",
                btn_access_wallet: "Open Wallet"
            },
            pt: {
                wallet_btn: "Minha Carteira",
                login_btn: "Entrar (Lojas)",
                register_btn: "Teste Grátis",
                hero_title: "Transforme clientes <br> em fãs.",
                hero_subtitle: "A plataforma de fidelização mais simples para o seu negócio.",
                stats_users: "Clientes Felizes",
                stats_stamps: "Selos Dados",
                stats_rewards: "Prémios Dados",
                modal_login_title: "Gestão de Loja",
                password: "Palavra-passe",
                btn_enter: "Entrar",
                new_business: "Novo negócio?",
                create_free_acc: "Criar Conta Grátis",
                modal_register_title: "Comece a Crescer",
                modal_register_sub: "30 dias grátis. Configuração em 1 minuto.",
                business_name: "Nome do Negócio",
                btn_launch: "Lançar Plataforma 🚀",
                has_account: "Já tem conta?",
                btn_login: "Entrar",
                wallet_title: "Minha Carteira",
                wallet_sub: "Aceda aos seus cartões com segurança.",
                btn_send_code: "Enviar Código",
                btn_access_wallet: "Abrir Carteira"
            }
        };

        const t = (key) => translations[lang.value][key] || key;

        // --- CONFIGURAÇÕES (THEME & LANG) ---
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

        // --- API CALLS ---
        const fetchConfig = async () => { try { const res = await fetch('/api/v1/system/config'); if(res.ok) storeConfig.value = await res.json(); } catch(e){} };
        const fetchStats = async () => { try { const res = await fetch('/api/v1/public/stats'); if(res.ok) stats.value = await res.json(); } catch(e){} };

        // --- LOGIN LOJA (COM DEBUG) ---
        const handleLogin = async () => {
            loading.value = true;
            errorMsg.value = '';
            
            try {
                // DEBUG: Para veres na consola o que está a ser enviado
                console.log("A tentar login com:", loginForm.value.email);

                const res = await fetch('/api/v1/auth/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        identifier: loginForm.value.email,
                        password: loginForm.value.password
                    })
                });

                if (res.ok) {
                    const data = await res.json();
                    window.location.href = data.redirect;
                } else {
                    const text = await res.text();
                    console.error("Erro Servidor:", text); // DEBUG: Mostra o erro real do Go
                    errorMsg.value = lang.value === 'pt' ? "Email ou Password incorretos." : "Invalid email or password.";
                }
            } catch (e) {
                console.error("Erro JS:", e); // DEBUG: Erros de rede ou JS
                errorMsg.value = lang.value === 'pt' ? "Erro de conexão." : "Connection error.";
            } finally {
                loading.value = false;
            }
        };
        
        // --- REGISTO NOVA LOJA ---
        const handleRegister = async () => {
            loading.value = true; 
            errorMsg.value = '';
            try {
                const res = await fetch('/api/v1/public/register', {
                    method: 'POST', 
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify(registerForm.value)
                });
                
                if (res.ok) {
                    const data = await res.json();
                    window.location.href = data.redirect;
                } else {
                    const txt = await res.text();
                    if(txt.includes("Email")) {
                        errorMsg.value = lang.value === 'pt' ? "Email já registado." : "Email already taken.";
                    } else {
                        errorMsg.value = lang.value === 'pt' ? "Erro ao criar conta." : "Error creating account.";
                    }
                }
            } catch (e) { 
                errorMsg.value = lang.value === 'pt' ? "Erro de conexão." : "Connection error."; 
            } finally { 
                loading.value = false; 
            }
        };

        // --- WALLET (SIMULAÇÃO 2FA) ---
        const requestWalletCode = () => {
            loading.value = true;
            // SIMULAÇÃO: No futuro isto chama o backend para enviar email
            setTimeout(() => {
                console.log("Code sent to", walletForm.value.email);
                loading.value = false;
                walletStep.value = 2; // Passa para o passo 2
                errorMsg.value = "";
            }, 1000);
        };

        const verifyWalletCode = () => {
            loading.value = true;
            // SIMULAÇÃO: Verifica se o código é 123456
            setTimeout(() => {
                if (walletForm.value.code === "123456") {
                    alert(lang.value === 'pt' ? "Acesso garantido! (Em desenvolvimento)" : "Access granted! (Dev mode)");
                    // window.location.href = '/wallet';
                } else {
                    errorMsg.value = lang.value === 'pt' ? "Código inválido." : "Invalid code.";
                }
                loading.value = false;
            }, 1000);
        };

        // --- INICIALIZAÇÃO ---
        onMounted(() => {
            initSettings();
            fetchConfig();
            fetchStats();
        });

        // --- EXPOR PARA O HTML ---
        return {
            theme, lang, currentModal, loading, errorMsg, storeConfig, stats,
            loginForm, registerForm, walletForm, walletStep,
            t, toggleTheme, toggleLang, 
            handleLogin, handleRegister, requestWalletCode, verifyWalletCode,
            openLogin: () => { currentModal.value = 'login'; errorMsg.value = ''; },
            openRegister: () => { currentModal.value = 'register'; errorMsg.value = ''; },
            openWallet: () => { currentModal.value = 'wallet'; walletStep.value = 1; errorMsg.value = ''; },
            closeModal: () => currentModal.value = null
        };
    }
};

Vue.createApp(LandingPage).mount('#app');