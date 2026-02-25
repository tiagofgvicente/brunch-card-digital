const LandingPage = {
    setup() {
        const { ref, computed, onMounted } = Vue;

        const theme = ref('light'); 
        const lang = ref('pt'); 
        const currentModal = ref(null); 
        const loading = ref(false);
        const errorMsg = ref('');
        const successMsg = ref(''); 
        
        const pendingRedirect = ref('');
        const storeTab = ref('login'); 
        const walletTab = ref('login'); 

        const forgotMode = ref(false);
        const forgotEmail = ref('');
        const resetToken = ref('');
        const resetType = ref(''); 
        const resetPasswordForm = ref({ password: '' });

        const billingCycle = ref('monthly');

        const storeConfig = ref({ name: 'Volto', logo: '' });
        const stats = ref({ total_cards: 0, total_stamps: 0, total_redeems: 0 });
        
        const loginForm = ref({ email: '', password: '' });
        const registerForm = ref({ name: '', email: '', password: '' });
        const walletForm = ref({ email: '', password: '' });
        const globalUserForm = ref({ firstName: '', lastName: '', email: '', phone: '', password: '', rgpd: false, marketing: false });

        const leadForm = ref({ businessType: '', contactName: '', email: '', phone: '' });
        const selectedTierForLead = ref('');

        const isWalletFormValid = computed(() => {
            const f = globalUserForm.value;
            return f.firstName.trim() !== '' && f.lastName.trim() !== '' &&
                   f.email.trim() !== '' && f.password.trim() !== '' && f.rgpd === true;
        });

        // 👇 AVALIAÇÃO EM TEMPO REAL DOS CAMPOS DA LEAD 👇
        const isLeadEmailInvalid = computed(() => {
            const email = leadForm.value.email;
            if (email.length === 0) return false; // Não mostra erro se estiver vazio
            return !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
        });

        const isLeadPhoneInvalid = computed(() => {
            const phone = leadForm.value.phone;
            if (phone.length === 0) return false; // Não mostra erro se estiver vazio
            const cleanPhone = phone.replace(/\s+/g, '');
            return !/^\+?[0-9]{9,}$/.test(cleanPhone);
        });

        const isLeadFormValid = computed(() => {
            const f = leadForm.value;
            const emailValid = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(f.email);
            const cleanPhone = f.phone.replace(/\s+/g, '');
            const phoneValid = /^\+?[0-9]{9,}$/.test(cleanPhone);
            
            return f.businessType.trim() !== '' && 
                   f.contactName.trim() !== '' && 
                   emailValid && 
                   phoneValid;
        });

        const translations = {
            en: {
                nav_wallet: "My Wallet", nav_business: "For Businesses", nav_pricing: "Pricing",
                tab_login: "Login", tab_register_store: "Start Free Trial", tab_register_wallet: "Create Account",
                btn_login: "Login", btn_enter: "Enter Dashboard",
                hero_title: "Turn customers <br> into regulars.", hero_subtitle: "The simplest loyalty platform for modern businesses.",
                stats_users: "Happy Users", stats_stamps: "Stamps Given", stats_rewards: "Rewards Claimed",
                modal_login_title: "Store Dashboard", password: "Password",
                modal_register_title: "Grow Your Business", modal_register_sub: "30 days free. Setup in 1 minute.",
                business_name: "Business Name", btn_launch: "Launch Platform",
                wallet_title: "My Wallet", wallet_sub: "Access all your cards securely.",
                forgot_pass: "Forgot your password?", forgot_title: "Recover Password", forgot_sub: "Enter your email. We'll send you a recovery link.",
                btn_send_link: "Send Link", back_login: "Back to Login", reset_title: "New Password", reset_sub: "Choose a secure new password.", btn_save_pass: "Save Password",
                
                price_title: "Simple & Transparent Pricing", price_sub: "Start with a 30-day free trial. No credit card required.",
                bill_monthly: "Monthly", bill_yearly: "Yearly",
                per_month: "/month", btn_start_trial: "Start Free Trial",
                btn_subscribe: "Subscribe Now",
                
                billed_monthly_note: "Billed monthly. Cancel anytime.",
                billed_yearly_basic: "Billed yearly at €180, cancel anytime.",
                billed_yearly_lite: "Billed yearly at €300, cancel anytime.",
                billed_yearly_pro: "Billed yearly at €360, cancel anytime.",
                save_tag: "Save",
                
                tier_basic: "Basic", tier_basic_sub: "Just the essentials to start",
                feat_basic_1: "Max 50 Customers", feat_basic_2: "1 Fixed standard stamp card", feat_basic_3: "No custom card design", feat_basic_4: "No Social Media or Maps", feat_basic_5: "No Customer Management", feat_basic_6: "No Reservations System",
                
                tier_lite: "Lite", tier_lite_sub: "Ideal for growing businesses",
                feat_lite_1: "Max 100 Customers", feat_lite_2: "1 Customizable Card (Icon OR Logo)", feat_lite_3: "Social Media & Maps in Wallet", feat_lite_4: "Basic Customer Management", feat_lite_5: "No Digital Menu", feat_lite_6: "No Reservations System",
                
                tier_pro: "Pro", tier_pro_sub: "Everything you need to scale",
                feat_pro_1: "Unlimited Customers & Advanced Mgmt", feat_pro_2: "Up to 3 Cards + Tiers (Bronze, Silver, Gold)", feat_pro_3: "1 Card 100% Custom Designed for You", feat_pro_4: "Both Logo AND Custom Icon", feat_pro_5: "Socials, Maps & Digital Menu in Wallet", feat_pro_6: "Reservations System (Brunch, Lunch, etc.)",
                
                foot_product: "Product", foot_resources: "Resources", foot_legal: "Legal",
                foot_features: "Features", foot_pricing: "Pricing", foot_login: "Login",
                foot_help: "Help Center", foot_contact: "Contact Us",
                foot_privacy: "Privacy Policy", foot_terms: "Terms of Service",
                
                // Erros inline
                lead_err_email: "Invalid email format.",
                lead_err_phone: "Must contain numbers only (min. 9 digits)."
            },
            pt: {
                nav_wallet: "Minha Carteira", nav_business: "Para Negócios", nav_pricing: "Preços",
                tab_login: "Entrar", tab_register_store: "Teste Grátis", tab_register_wallet: "Criar Conta",
                btn_login: "Entrar", btn_enter: "Entrar no Painel",
                hero_title: "Transforme clientes <br> em fãs.", hero_subtitle: "A plataforma de fidelização mais simples para o seu negócio.",
                stats_users: "Clientes Felizes", stats_stamps: "Selos Dados", stats_rewards: "Prémios Dados",
                modal_login_title: "Gestão de Loja", password: "Palavra-passe",
                modal_register_title: "Comece a Crescer", modal_register_sub: "30 dias grátis. Configuração em 1 minuto.",
                business_name: "Nome do Negócio", btn_launch: "Lançar Plataforma",
                wallet_title: "Minha Carteira", wallet_sub: "Aceda aos seus cartões com segurança.",
                forgot_pass: "Esqueceu-se da palavra-passe?", forgot_title: "Recuperar Palavra-passe", forgot_sub: "Insira o seu email. Vamos enviar-lhe um link de recuperação.",
                btn_send_link: "Enviar Link", back_login: "Voltar ao Login", reset_title: "Nova Palavra-passe", reset_sub: "Escolha uma nova palavra-passe segura.", btn_save_pass: "Guardar Palavra-passe",
                
                price_title: "Planos Simples e Transparentes", price_sub: "Comece com 30 dias grátis. Sem cartão de crédito.",
                bill_monthly: "Mensal", bill_yearly: "Anual",
                per_month: "/mês", btn_start_trial: "Começar Teste Grátis",
                btn_subscribe: "Subscrever Agora", 
                
                billed_monthly_note: "Faturado mensalmente. Cancele quando quiser.",
                billed_yearly_basic: "Faturado anualmente a €180. Cancele quando quiser.",
                billed_yearly_lite: "Faturado anualmente a €300. Cancele quando quiser.",
                billed_yearly_pro: "Faturado anualmente a €360. Cancele quando quiser.",
                save_tag: "Poupança de",
                
                tier_basic: "Basic", tier_basic_sub: "O essencial para arrancar",
                feat_basic_1: "Máximo 50 Clientes", feat_basic_2: "1 Cartão (Selo Fixo Standard)", feat_basic_3: "Sem design personalizável", feat_basic_4: "Sem Redes Sociais, Menu ou Mapas", feat_basic_5: "Sem Gestão de Clientes", feat_basic_6: "Sem Sistema de Reservas",
                
                tier_lite: "Lite", tier_lite_sub: "Ideal para pequenos negócios",
                feat_lite_1: "Máximo 100 Clientes", feat_lite_2: "1 Cartão Personalizável (Ícone OU Logo)", feat_lite_3: "Redes Sociais e Mapas na Wallet", feat_lite_4: "Gestão Básica de Clientes", feat_lite_5: "Sem Menu Digital", feat_lite_6: "Sem Sistema de Reservas",
                
                tier_pro: "Pro", tier_pro_sub: "Tudo o que o negócio precisa",
                feat_pro_1: "Clientes Ilimitados e Gestão Avançada", feat_pro_2: "Até 3 Cartões + Níveis (Bronze, Silver, Gold)", feat_pro_3: "1 Cartão 100% Desenhado à Imagem da Loja", feat_pro_4: "Ícone e Logótipo em simultâneo", feat_pro_5: "Redes Sociais, Mapas e Menu Digital", feat_pro_6: "Sistema de Reservas (Brunch, Almoço, etc.)",
                
                foot_product: "Produto", foot_resources: "Recursos", foot_legal: "Legal",
                foot_features: "Funcionalidades", foot_pricing: "Preços", foot_login: "Entrar na Loja",
                foot_help: "Centro de Ajuda", foot_contact: "Contacte-nos",
                foot_privacy: "Política de Privacidade", foot_terms: "Termos de Serviço",
                
                // Erros inline
                lead_err_email: "Formato de email inválido.",
                lead_err_phone: "Tem de conter apenas números (mín. 9)."
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

        const openSubscribeModal = (tier) => {
            selectedTierForLead.value = tier;
            currentModal.value = 'lead_capture';
            errorMsg.value = '';
            successMsg.value = '';
        };

        const submitLeadForm = async () => {
            if (!isLeadFormValid.value) return; // Segurança extra

            loading.value = true;
            errorMsg.value = '';
            successMsg.value = '';
            
            const cleanPhone = leadForm.value.phone.replace(/\s+/g, '');

            try {
                const payload = {
                    company_name: leadForm.value.businessType, 
                    contact_name: leadForm.value.contactName, 
                    email: leadForm.value.email,
                    phone: cleanPhone, 
                    tier: selectedTierForLead.value,
                    cycle: billingCycle.value,
                    lang: lang.value
                };

                const res = await fetch('/api/v1/public/capture-lead', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                });

                if (res.ok) {
                    successMsg.value = lang.value === 'pt' 
                        ? "Recebemos o seu pedido! Um dos nossos agentes vai contactá-lo muito em breve para o ajudar." 
                        : "Request received! An agent will contact you shortly to help you setup.";
                    leadForm.value = { businessType: '', contactName: '', email: '', phone: '' };
                } else {
                    const errorText = await res.text();
                    errorMsg.value = errorText || (lang.value === 'pt' ? "Erro ao enviar pedido. Tente novamente." : "Error sending request. Please try again.");
                }
            } catch(e) {
                errorMsg.value = lang.value === 'pt' ? "Erro de ligação ao servidor." : "Connection error to server.";
            } finally {
                loading.value = false;
            }
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

        const scrollToPricing = () => {
            const el = document.getElementById('pricing');
            if(el) el.scrollIntoView({ behavior: 'smooth' });
        };

        onMounted(() => { 
            initSettings(); fetchConfig(); fetchStats(); 
            
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
            isWalletFormValid, isLeadFormValid, isLeadEmailInvalid, isLeadPhoneInvalid, t, toggleTheme, toggleLang, 
            handleLogin, handleRegister, handleWalletRegister, handleWalletLogin, proceedToApp,
            forgotMode, forgotEmail, resetToken, resetType, resetPasswordForm, handleForgotPassword, handleResetPassword,
            
            billingCycle, scrollToPricing, 
            
            leadForm, openSubscribeModal, submitLeadForm,
            
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