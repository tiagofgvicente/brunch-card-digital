const AdminApp = {
    setup() {
        const { ref, onMounted, nextTick, watch, computed } = Vue;

        // --- STATE ---
        const urlParams = new URLSearchParams(window.location.search);
        const currentStore = urlParams.get('store');
        const api = (url) => currentStore ? `${url}${url.includes('?') ? '&' : '?'}store=${currentStore}` : url;

        const cards = ref([]);
        const currentPage = ref('customers');
        const searchQuery = ref('');
        const activeSkin = ref('default');
        const showPreview = ref(false);
        const previewDesign = ref('default');
        const toasts = ref([]);
        const lang = ref('pt'); // Default
        
        const showIconPicker = ref(false);
        const icons = ['🍳','🍔','🍕','🌭','🥪','🌮','🥗','🥐','🥯','🥞','☕','🍺','🍷','🍹','🥤','🧋','🍩','🍪','🍰','🍦','✂️','💇','💅','💄','💈','🏋️','🧘','🚗','🚲','🎮','🐾','🐶','🐱','📚','💊','🦷','🕶️','💍','👠','🧢'];

        // ADICIONEI OS NOVOS CAMPOS AO DEFAULT STATE
        const storeConfig = ref({ 
            name: 'Store', logo_url: '', themeMode: 'dark', 
            primary_color: '#00a896', 
            text_color: '#ffffff', // Novo
            border_color: '#ffffff', // Novo
            card_image_url: '', // Novo
            stamp_icon: '🍳',
            bronzeThreshold: 15, silverThreshold: 40, goldThreshold: 100 
        });
        
        const newMember = ref({ first_name: '', last_name: '', email: '', phone: '', rgpd: false, marketing: false });
        const editingCard = ref(null);
        const lastScanStatus = ref('');
        const scannerInput = ref(null);
        const availableSkins = ref([]);

        // --- CUSTOM DESIGNER STATE ---
        const showCustomDesigner = ref(false);
        const designForm = ref({
            background: '#00a896',
            textColor: '#ffffff',
            borderColor: '#ffffff',
            image: null,
            zoom: 100
        });

        // --- TRADUÇÕES ---
        const translations = {
            en: {
                nav_customers: "Customers", nav_register: "New Member", nav_scanner: "Scan Station", nav_skins: "Card Skins", nav_settings: "Settings", nav_logout: "Logout",
                title_customers: "Customer Management", title_register: "Register New Member", title_scanner: "Ready to Scan", title_skins: "Card Skins", title_settings: "System Settings",
                sub_skins: "Choose the visual style for your customers' digital cards.", sub_settings: "Manage your store identity and dashboard appearance.", sub_scanner: "Use the QR Reader. The stamp will be applied automatically.",
                th_name: "Name & Contact", th_status: "Status", th_consent: "Consent", th_avail: "Available", th_redeem: "Redeemed", th_actions: "Actions",
                sec_brand: "Branding & Identity", lbl_store_name: "Store Name", lbl_color: "Primary Brand Color", lbl_icon: "Stamp Icon", lbl_logo: "Store Logo", btn_upload: "UPLOAD NEW IMAGE",
                sec_tiers: "Loyalty Tiers", sec_security: "Security", lbl_pass_key: "Staff Access Key", btn_save_config: "SAVE CONFIGURATION", btn_change_pass: "Update Password",
                lbl_fname: "First Name", lbl_lname: "Last Name", lbl_email: "Email", lbl_phone: "Phone", lbl_rgpd: "Client accepts Privacy Policy", lbl_mkt: "Receive offers via email", btn_create: "CREATE CARD",
                search_placeholder: "Search customer...", active_badge: "ACTIVE", btn_preview: "Preview", btn_activate: "ACTIVATE",
                msg_no_email: "Without email, the customer won't receive the digital card or rewards. They will only be registered in the system."
            },
            pt: {
                nav_customers: "Clientes", nav_register: "Novo Membro", nav_scanner: "Scanner", nav_skins: "Estilo Cartão", nav_settings: "Definições", nav_logout: "Sair",
                title_customers: "Gestão de Clientes", title_register: "Registar Novo Membro", title_scanner: "Pronto a Ler", title_skins: "Estilo do Cartão", title_settings: "Definições de Sistema",
                sub_skins: "Escolha o visual do cartão digital dos seus clientes.", sub_settings: "Gerira a identidade e aparência da loja.", sub_scanner: "Use o leitor QR. O selo será aplicado automaticamente.",
                th_name: "Nome & Contacto", th_status: "Estado", th_consent: "Privacidade", th_avail: "Disponível", th_redeem: "Usados", th_actions: "Ações",
                sec_brand: "Marca & Identidade", lbl_store_name: "Nome da Loja", lbl_color: "Cor Principal", lbl_icon: "Ícone Selo", lbl_logo: "Logótipo", btn_upload: "CARREGAR IMAGEM",
                sec_tiers: "Níveis de Fidelidade", sec_security: "Segurança", lbl_pass_key: "Password de Acesso", btn_save_config: "GUARDAR CONFIGURAÇÃO", btn_change_pass: "Atualizar Password",
                lbl_fname: "Primeiro Nome", lbl_lname: "Último Nome", lbl_email: "Email", lbl_phone: "Telemóvel", lbl_rgpd: "Cliente aceita Política de Privacidade", lbl_mkt: "Receber ofertas por email", btn_create: "CRIAR CARTÃO",
                search_placeholder: "Procurar cliente...", active_badge: "ATIVO", btn_preview: "Ver", btn_activate: "ATIVAR",
                msg_no_email: "Sem email, o cliente não recebe o cartão digital nem prémios. Apenas fica registado no sistema."
            }
        };

        const t = (key) => translations[lang.value][key] || key;

        // --- ACTIONS ---
        const setTheme = (m) => { 
            storeConfig.value.themeMode = m; 
            document.documentElement.setAttribute('data-theme', m);
        };

        const toggleTheme = () => {
            const newTheme = storeConfig.value.themeMode === 'dark' ? 'light' : 'dark';
            setTheme(newTheme);
            localStorage.setItem('theme', newTheme);
        };

        const toggleLang = () => {
            lang.value = lang.value === 'pt' ? 'en' : 'pt';
            localStorage.setItem('lang', lang.value);
        };

        const initSettings = () => {
            const savedLang = localStorage.getItem('lang');
            if (savedLang) lang.value = savedLang;

            const savedTheme = localStorage.getItem('theme');
            if (savedTheme) {
                storeConfig.value.themeMode = savedTheme;
                document.documentElement.setAttribute('data-theme', savedTheme);
            }
        };

        const showToast = (message, type = 'success') => {
            const id = Date.now();
            toasts.value.push({ id, message, type });
            setTimeout(() => { toasts.value = toasts.value.filter(t => t.id !== id); }, 3000);
        };

        // --- COMPUTED PREVIEW ---
        const previewColors = computed(() => {
            const skinId = previewDesign.value;
            
            // Se for custom, usa as cores da config
            if (skinId === 'custom') {
                return { 
                    bg: storeConfig.value.card_image_url ? `url(${storeConfig.value.card_image_url})` : (storeConfig.value.primary_color || '#00a896'), 
                    text: storeConfig.value.text_color || '#ffffff' 
                };
            }
            
            const skin = availableSkins.value.find(s => s.id === skinId);
            if (skin) {
                if (skin.image) return { bg: `url(${skin.image})`, text: '#ffffff' };
                if (skin.style) {
                     const bg = skin.style.includes('background:') ? skin.style.split('background:')[1].split(';')[0] : '#333';
                     return { bg: bg, text: '#ffffff' };
                }
            }
            // Fallback para Default (agora isolado)
            if (skinId === 'black') return { bg: '#1a1a1a', text: '#ffd166' };
            if (skinId === 'gold') return { bg: 'linear-gradient(45deg, #FFD700, #FDB931)', text: '#000000' };
            
            // DEFAULT PURO
            return { bg: '#00a896', text: '#ffffff' };
        });

        const filteredCards = computed(() => {
            if (!searchQuery.value) return cards.value;
            const q = searchQuery.value.toLowerCase();
            return cards.value.filter(c => c.customer_id.toLowerCase().includes(q) || c.last_name.toLowerCase().includes(q) || (c.email && c.email.toLowerCase().includes(q)) || (c.phone && c.phone.includes(q)));
        });

        // --- API CALLS ---
        const fetchCards = async () => { 
            try { const res = await fetch(api('/api/v1/admin/cards')); if (res.ok) cards.value = await res.json(); } catch(e) { console.error(e); }
        };
        
        const fetchAvailableSkins = async () => {
            try {
                const res = await fetch(api('/api/v1/system/skins'));
                if(res.ok) {
                    const globals = await res.json();
                    const customSkin = { id: 'custom', name: 'Custom Brand', type: 'standard', style: `background: ${storeConfig.value.primary_color || '#00a896'}` };
                    availableSkins.value = [customSkin, ...globals];
                }
            } catch(e) { console.error(e); }
        };

        const getPreviewStyle = (skin) => {
            if (!skin) return {};
            if (skin.image) return { backgroundImage: `url(${skin.image})`, backgroundSize: 'cover', backgroundPosition: 'center' };
            // FIX: Default agora é sempre estático
            if(skin.id === 'default') return { background: '#00a896', color: '#fff' };
            
            if(skin.id === 'black') return { background: '#1a1a1a', borderBottom: '3px solid #ffd166', color: '#fff' };
            if(skin.id === 'custom') return { backgroundColor: storeConfig.value.primary_color || '#00a896', color: '#fff' };
            if(skin.style) return { background: skin.style.replace('background:', '').replace(';', ''), color: '#fff' };
            return { background: '#ccc' };
        };

        const changePage = (page) => {
            currentPage.value = page;
            if (page === 'scanner') nextTick(() => focusScanner());
            if (page !== 'settings' && page !== 'skins') fetchCards();
        };

        const selectIcon = (icon) => { storeConfig.value.stamp_icon = icon; showIconPicker.value = false; };

        const saveSettings = async () => { 
            if (activeSkin.value !== 'custom') { await updateGlobalSkin('custom'); }
            await fetch(api('/api/v1/admin/settings'), { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(storeConfig.value) }); 
            showToast("Configuration Saved!", "success"); 
            fetch(api('/api/v1/system/config')).then(r => r.json()).then(d => { 
                storeConfig.value = { ...storeConfig.value, ...d }; 
                fetchAvailableSkins();
                document.title = `Volto Store Admin | ${d.name}`;
            });
        };

        // --- CUSTOM DESIGNER LOGIC ---
        const openCustomDesigner = () => {
            // Inicializar com os valores atuais do storeConfig
            designForm.value = {
                background: storeConfig.value.primary_color || '#00a896',
                textColor: storeConfig.value.text_color || '#ffffff', 
                borderColor: storeConfig.value.border_color || '#ffffff',
                image: storeConfig.value.card_image_url || null,
                zoom: 100
            };
            showCustomDesigner.value = true;
        };

        const handleBgUpload = (e) => {
            const file = e.target.files[0];
            if (!file) return;
            const reader = new FileReader();
            reader.onload = (ev) => { designForm.value.image = ev.target.result; };
            reader.readAsDataURL(file);
        };

        const saveCustomDesign = async () => {
            // 1. Atualizar o config local com TODAS as propriedades
            storeConfig.value.primary_color = designForm.value.background;
            storeConfig.value.text_color = designForm.value.textColor;
            storeConfig.value.border_color = designForm.value.borderColor;
            if(designForm.value.image) {
                storeConfig.value.card_image_url = designForm.value.image;
            }

            // 2. Enviar para a API
            await fetch(api('/api/v1/admin/settings'), { 
                method: 'POST', 
                headers: { 'Content-Type': 'application/json' }, 
                body: JSON.stringify(storeConfig.value) 
            });
            
            showToast("Custom Brand Updated!", "success");
            showCustomDesigner.value = false;
            fetchAvailableSkins();
        };

        const handleLogoUpload = (e) => { const r=new FileReader(); r.onload=(ev)=>{ storeConfig.value.logo_url=ev.target.result; }; r.readAsDataURL(e.target.files[0]); };
        const removeLogo = () => { storeConfig.value.logo_url=''; };

        const updatePassword = async () => { 
            const o=document.getElementById('pass-old').value, n=document.getElementById('pass-new').value; 
            const res = await fetch(api('/api/v1/admin/update-password'), {method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({old:o, new:n})}); 
            if(res.ok) showToast("Password Updated!", "success"); else showToast("Error updating password", "error");
        };

        const updateGlobalSkin = async (s) => { 
            await fetch(api('/api/v1/admin/update-skin'), {method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({design:s})}); 
            activeSkin.value = s; showToast("Skin Activated!", "success");
        };
        const openPreview = (d) => { previewDesign.value = d; showPreview.value = true; };
        const openEditModal = (c) => editingCard.value = { ...c };
        const saveEdit = async () => { await fetch(api('/api/v1/admin/update'), { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(editingCard.value) }); showToast("Customer Updated!", "success"); editingCard.value = null; fetchCards(); };
        const toggleConsent = async (c, type) => { const f = type === 'rgpd' ? 'rgpd_accepted' : 'marketing_accepted'; c[f] = !c[f]; await fetch(api('/api/v1/admin/update-consent'), { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ id: c.id, [f]: c[f] }) }); };
        const registerMember = async () => { const res = await fetch(api('/api/v1/cards'), { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ customer_id: newMember.value.first_name, last_name: newMember.value.last_name, email: newMember.value.email, phone: newMember.value.phone, rgpd_accepted: newMember.value.rgpd, marketing_accepted: newMember.value.marketing }) }); if(res.ok) { showToast("Member Registered!", "success"); newMember.value = {first_name:'', last_name:'', email:'', phone:'', rgpd:false, marketing:false}; changePage('customers'); } };
        const logout = async () => { await fetch(api('/api/v1/auth/logout'), { method: 'POST' }); window.location.href = "/"; };
        const focusScanner = () => { if(scannerInput.value) scannerInput.value.focus(); };
        const handleScan = async (e) => { let id = e.target.value; e.target.value = ''; if(id.includes('id=')) id = id.split('id=')[1].split('&')[0]; if(id.length>5) { const res = await fetch(api(`/api/v1/cards/stamp?id=${id}`), { method: 'POST' }); if(res.ok) { const d = await res.json(); lastScanStatus.value = `Success: ${d.customer_id}`; setTimeout(()=>lastScanStatus.value='',3000); } else lastScanStatus.value = "Error"; } };
        const calculateAvailable = (c) => Math.max(0, Math.floor(c.total_stamps / 10) - (c.total_redeemed_bonuses || 0));
        const addStampFromAdmin = async (c) => { await fetch(api(`/api/v1/cards/stamp?id=${c.id}`), {method:'POST'}); fetchCards(); showToast("Stamp Added", "success"); };
        const redeemFromAdmin = async (c) => { await fetch(api(`/api/v1/cards/use-reward?id=${c.id}`), {method:'POST'}); fetchCards(); showToast("Reward Redeemed", "success"); };
        const resetCard = async (id) => { if(confirm("Reset?")) { await fetch(api(`/api/v1/admin/reset?id=${id}`), {method:'POST'}); fetchCards(); showToast("Card Reset", "warning"); } };
        const viewCard = (id) => window.open(`/card?store=${currentStore}&id=${id}`, '_blank');
        
        // --- INIT ---
        onMounted(() => { 
            initSettings(); // 1. Carrega as preferências do browser
            fetchCards(); 
            
            fetch(api('/api/v1/system/config')).then(r => r.json()).then(d => { 
                storeConfig.value = { ...storeConfig.value, ...d }; 
                activeSkin.value = d.card_skin || 'default';
                
                // 2. Lógica de prioridade:
                const localTheme = localStorage.getItem('theme');
                if (localTheme) {
                    setTheme(localTheme);
                } else {
                    setTheme(d.themeMode || 'dark');
                }
                
                fetchAvailableSkins();
                document.title = `Volto Store Admin | ${d.name}`;
            }); 
        });

        return { 
            cards, currentPage, searchQuery, filteredCards, storeConfig, 
            newMember, editingCard, lastScanStatus, scannerInput, activeSkin,
            showIconPicker, icons, selectIcon, saveSettings, handleLogoUpload, removeLogo,
            changePage, logout, registerMember, focusScanner, handleScan, openEditModal, saveEdit, toggleConsent,
            calculateAvailable, addStampFromAdmin, redeemFromAdmin, viewCard, resetCard, updatePassword, setTheme, toggleTheme,
            updateGlobalSkin, openPreview, showPreview, previewColors, availableSkins, getPreviewStyle,
            toasts, showToast, lang, toggleLang, t,
            // Exports do Designer
            showCustomDesigner, designForm, openCustomDesigner, handleBgUpload, saveCustomDesign
        };
    }
};

Vue.createApp(AdminApp).mount('#app');