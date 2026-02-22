const AdminApp = {
    setup() {
        const { ref, onMounted, nextTick, computed } = Vue;

        const urlParams = new URLSearchParams(window.location.search);
        const currentStore = urlParams.get('store');
        const api = (url) => currentStore ? `${url}${url.includes('?') ? '&' : '?'}store=${currentStore}` : url;

        const cards = ref([]);
        const currentPage = ref('customers');
        const searchQuery = ref('');
        const activeTabScope = ref('all');
        const activeSkin = ref('default');
        const showPreview = ref(false);
        const previewDesign = ref('default');
        const toasts = ref([]);
        const lang = ref('pt');
        
        const showIconPicker = ref(false);
        const iconPickerTarget = ref(null); // 'new' ou 'edit'
        const icons = ['🍳','🍔','🍕','🌭','🥪','🌮','🥗','🥐','🥯','🥞','☕','🍺','🍷','🍹','🥤','🧋','🍩','🍪','🍰','🍦','✂️','💇','💅','💄','💈','🏋️','🧘','🚗','🚲','🎮','🐾','🐶','🐱','📚','💊','🦷','🕶️','💍','👠','🧢', '💳', '🎁', '🛒'];

        const storeConfig = ref({ 
            name: 'Store', logo_url: '', themeMode: 'dark', 
            primary_color: '#00a896', text_color: '#ffffff', border_color: '#ffffff',    
            card_image_url: '', card_image_zoom: 100, card_image_pos_x: 0, card_image_pos_y: 0,
            social_instagram: '', social_facebook: '', social_twitter: '', social_whatsapp: '', social_tiktok: '', social_youtube: '', social_website: '',
            menu_url: '', location_url: '', 
            bronzeThreshold: 15, silverThreshold: 40, goldThreshold: 100, tier: 'pro'
        });

        // --- GESTÃO DE ÂMBITOS (SCOPES) ---
        const storeScopes = ref([]); 
        const activeScannerScope = ref('');
        
        // A LISTA DEFINIDA POR TI
        const predefinedScopeNames = ['Geral', 'Pequeno-Almoço', 'Menu de Almoço', 'Sopa', 'Lanche', 'Jantar', 'Cafetaria', 'Bebidas'];
        
        const newScopeName = ref(predefinedScopeNames[7]); // Bebidas default
        const newScopeIcon = ref('🍹');

        const editingScopeId = ref(null);
        const editScopeTempName = ref('');
        const editScopeTempIcon = ref('');
        
        const fetchScopes = async () => {
            try {
                const res = await fetch(api('/api/v1/admin/scopes'), { headers: { 'Cache-Control': 'no-cache' } });
                if (res.ok) {
                    storeScopes.value = await res.json();
                    const activeScopes = storeScopes.value.filter(s => s.is_active);
                    if (activeScopes.length > 0 && !activeScannerScope.value) {
                        activeScannerScope.value = activeScopes[0].id;
                    }
                }
            } catch(e) {}
        };

        const createScope = async () => {
            if(!newScopeName.value) return;
            const res = await fetch(api('/api/v1/admin/scopes/create'), {
                method: 'POST', headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name: newScopeName.value, stamp_icon: newScopeIcon.value })
            });
            if(res.ok) {
                showToast("Novo cartão criado!", "success");
                newScopeName.value = predefinedScopeNames[7];
                fetchScopes();
            } else showToast("Erro (Âmbito já existe?)", "error");
        };

        const startEditScope = (scope) => {
            editingScopeId.value = scope.id;
            editScopeTempName.value = scope.name;
            editScopeTempIcon.value = scope.stamp_icon;
        };

        const cancelEditScope = () => { editingScopeId.value = null; };

        const saveEditScope = async () => {
            const res = await fetch(api('/api/v1/admin/scopes/update'), {
                method: 'POST', headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ id: editingScopeId.value, name: editScopeTempName.value, stamp_icon: editScopeTempIcon.value })
            });
            if(res.ok) {
                showToast("Cartão atualizado!", "success");
                editingScopeId.value = null;
                fetchScopes();
            } else showToast("Erro ao atualizar", "error");
        };

        const toggleScope = async (scope) => {
            if(scope.is_main) { showToast("Não pode desativar o cartão principal", "error"); return; }
            const res = await fetch(api('/api/v1/admin/scopes/toggle'), {
                method: 'POST', headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ id: scope.id, is_active: !scope.is_active })
            });
            if(res.ok) { showToast("Status atualizado!", "success"); fetchScopes(); }
        };

        const deleteScope = async (scope) => {
            if (scope.is_main) { showToast("Não pode apagar o cartão principal", "error"); return; }
            
            // Alerta de segurança
            if (!confirm(`⚠️ ATENÇÃO!\nTem a certeza que quer APAGAR o cartão "${scope.name}"?\nTodos os clientes que têm este cartão vão perdê-lo permanentemente.`)) return;

            const res = await fetch(api('/api/v1/admin/scopes/delete'), {
                method: 'POST', headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ id: scope.id })
            });
            
            if (res.ok) {
                showToast("Cartão apagado com sucesso!", "success");
                fetchScopes(); // Atualiza a lista de âmbitos
                fetchCards();  // Atualiza a tabela de clientes para remover os cartões apagados
            } else {
                showToast("Erro ao apagar cartão.", "error");
            }
        };

        const openIconPicker = (target) => {
            iconPickerTarget.value = target;
            showIconPicker.value = true;
        };

        const selectIcon = (icon) => { 
            if (iconPickerTarget.value === 'new') newScopeIcon.value = icon;
            if (iconPickerTarget.value === 'edit') editScopeTempIcon.value = icon;
            showIconPicker.value = false; 
        };
        // ------------------------------------------

        const settingsForm = ref({});
        const newMember = ref({ first_name: '', last_name: '', email: '', phone: '', rgpd: false, marketing: false });
        const editingCard = ref(null);
        const availableSkins = ref([]); 

        const lastScanStatus = ref('');
        let html5Qrcode = null;
        let isScanning = false;

        const startScanner = () => {
            nextTick(() => {
                if (!html5Qrcode) html5Qrcode = new Html5Qrcode("reader");
                if (!html5Qrcode.isScanning) {
                    html5Qrcode.start({ facingMode: "environment" }, { fps: 10, qrbox: { width: 250, height: 250 } }, onScanSuccess, () => {}).catch(err => {
                        lastScanStatus.value = "⚠️ Permita o acesso à câmara.";
                    });
                }
            });
        };

        const stopScanner = () => {
            if (html5Qrcode && html5Qrcode.isScanning) html5Qrcode.stop().then(() => html5Qrcode.clear()).catch(e => console.error(e));
        };

        const onScanSuccess = async (decodedText) => {
            if (isScanning) return; 
            isScanning = true;

            try {
                const data = JSON.parse(decodedText);
                
                if (data && data.action === 'global_register') {
                    if (!activeScannerScope.value) {
                        lastScanStatus.value = "❌ Erro: Selecione um Âmbito (Cartão) primeiro!";
                        showToast("Selecione um cartão para atribuir.", "error");
                        setTimeout(() => { isScanning = false; }, 2000);
                        return;
                    }

                    const res = await fetch(api('/api/v1/cards'), { 
                        method: 'POST', headers: { 'Content-Type': 'application/json' }, 
                        body: JSON.stringify({ 
                            customer_id: data.first || 'Novo', last_name: data.last || 'Cliente', 
                            email: data.email || '', phone: data.phone || '', 
                            scope_id: activeScannerScope.value, 
                            rgpd_accepted: true, marketing_accepted: false 
                        }) 
                    });
                    if(res.ok) {
                        lastScanStatus.value = `✅ Novo cliente registado no cartão selecionado!`;
                        showToast(`Cliente registado com sucesso`, "success");
                        await fetchCards();
                    } else lastScanStatus.value = "❌ O cliente já tem este cartão nesta loja.";
                } 
                else if (data && data.action === 'redeem') {
                    if (data.store !== currentStore) {
                        lastScanStatus.value = "❌ Erro: Este prémio pertence a outra loja!";
                    } else {
                        const res = await fetch(api(`/api/v1/cards/use-reward?id=${data.id}`), { method: 'POST' });
                        if(res.ok) {
                            lastScanStatus.value = `🎁 Prémio descontado com sucesso!`;
                            await fetchCards();
                        } else lastScanStatus.value = "❌ Erro ao validar prémio.";
                    }
                }
            } catch (error) {
                let id = decodedText;
                let storeInQr = '';

                if (id.includes('store=')) storeInQr = id.split('store=')[1].split('&')[0];
                if (id.includes('id=')) id = id.split('id=')[1].split('&')[0];

                if (storeInQr && storeInQr !== currentStore) {
                    lastScanStatus.value = "❌ Erro: Cartão pertence a outra loja!";
                } 
                else if(id.length > 5 && !id.includes('{')) {
                    try {
                        const res = await fetch(api(`/api/v1/cards/stamp?id=${id}`), { method: 'POST' });
                        if(res.ok) {
                            const d = await res.json();
                            lastScanStatus.value = `✅ Selo adicionado a ${d.customer_id}!`;
                            await fetchCards(); 
                        } else lastScanStatus.value = "❌ Erro ao adicionar selo.";
                    } catch(e) { lastScanStatus.value = "❌ Erro de ligação."; }
                } else lastScanStatus.value = "❌ QR Code Inválido.";
            }

            setTimeout(() => { lastScanStatus.value = ''; isScanning = false; }, 3000);
        };

        const isDragging = ref(false);
        const dragStart = ref({ x: 0, y: 0 });
        const showCustomDesigner = ref(false);
        const designForm = ref({ background: '#00a896', textColor: '#ffffff', borderColor: '#ffffff', image: null, zoom: 100, posX: 0, posY: 0 });

        const startDrag = (e) => { if (!designForm.value.image) return; isDragging.value = true; dragStart.value.x = e.clientX - designForm.value.posX; dragStart.value.y = e.clientY - designForm.value.posY; };
        const onDrag = (e) => { if (!isDragging.value) return; designForm.value.posX = e.clientX - dragStart.value.x; designForm.value.posY = e.clientY - dragStart.value.y; };
        const stopDrag = () => { isDragging.value = false; };

        const handleMenuUpload = (e) => { 
            const file = e.target.files[0]; 
            if (!file) return; 
            if (file.type !== 'application/pdf') { showToast("Apenas ficheiros PDF são permitidos.", "error"); return; }
            if (file.size > 5 * 1024 * 1024) { showToast("O PDF é demasiado grande. Máximo 5MB.", "error"); return; }
            const reader = new FileReader(); 
            reader.onload = (ev) => { settingsForm.value.menu_url = ev.target.result; }; 
            reader.readAsDataURL(file); 
        };
        const removeMenu = () => { settingsForm.value.menu_url = ''; };

        const translations = {
            en: { 
                nav_customers: "Customers", nav_register: "New Member", nav_scanner: "Scan Station", nav_skins: "Card Skins", nav_settings: "Settings", nav_logout: "Logout", 
                title_customers: "Customer Management", title_register: "Register New Member", title_scanner: "Ready to Scan", title_skins: "Card Skins", title_settings: "System Settings", 
                sub_skins: "Choose the visual style for your customers' digital cards.", sub_settings: "Manage your store identity and dashboard appearance.", sub_scanner: "Use the QR Reader. The stamp will be applied automatically.", 
                th_name: "Name & Contact", th_status: "Status", th_consent: "Consent", th_avail: "Available", th_redeem: "Redeemed", th_actions: "Actions", 
                sec_brand: "Branding & Identity", lbl_store_name: "Store Name", lbl_color: "Primary Brand Color", lbl_logo: "Store Logo", btn_upload: "UPLOAD NEW IMAGE", 
                sec_tiers: "Loyalty Tiers", sec_security: "Security", lbl_pass_key: "Dashboard Login Password", sub_pass_key: "The password you use to log into this management platform.", btn_save_config: "SAVE CONFIGURATION", btn_change_pass: "Update Password", 
                lbl_fname: "First Name", lbl_lname: "Last Name", lbl_email: "Email", lbl_phone: "Phone", lbl_rgpd: "Client accepts Privacy Policy", lbl_mkt: "Receive offers via email", btn_create: "CREATE CARD", 
                search_placeholder: "Search customer...", active_badge: "ACTIVE", btn_preview: "Preview", btn_activate: "ACTIVATE", 
                msg_no_email: "Without email, the customer won't receive the digital card or rewards. They will only be registered in the system." 
            },
            pt: { 
                nav_customers: "Clientes", nav_register: "Novo Membro", nav_scanner: "Scanner", nav_skins: "Estilo Cartão", nav_settings: "Definições", nav_logout: "Sair", 
                title_customers: "Gestão de Clientes", title_register: "Registar Novo Membro", title_scanner: "Pronto a Ler", title_skins: "Estilo do Cartão", title_settings: "Definições de Sistema", 
                sub_skins: "Escolha o visual do cartão digital dos seus clientes.", sub_settings: "Gerira a identidade e aparência da loja.", sub_scanner: "Aponte a câmara para o QRCode do cliente.", 
                th_name: "Nome & Contacto", th_status: "Estado", th_consent: "Privacidade", th_avail: "Disponível", th_redeem: "Usados", th_actions: "Ações", 
                sec_brand: "Marca & Identidade", lbl_store_name: "Nome da Loja", lbl_color: "Cor Principal", lbl_logo: "Logótipo", btn_upload: "CARREGAR IMAGEM", 
                sec_tiers: "Níveis de Fidelidade", sec_security: "Segurança", lbl_pass_key: "Password de Acesso (Login)", sub_pass_key: "A password que utiliza para entrar nesta plataforma de gestão.", btn_save_config: "GUARDAR CONFIGURAÇÃO", btn_change_pass: "Atualizar Password", 
                lbl_fname: "Primeiro Nome", lbl_lname: "Último Nome", lbl_email: "Email", lbl_phone: "Telemóvel", lbl_rgpd: "Cliente aceita Política de Privacidade", lbl_mkt: "Receber ofertas por email", btn_create: "CRIAR CARTÃO", 
                search_placeholder: "Procurar cliente...", active_badge: "ATIVO", btn_preview: "Ver", btn_activate: "ATIVAR", 
                msg_no_email: "Sem email, o cliente não recebe o cartão digital nem prémios. Apenas fica registado no sistema." 
            }
        };
        const t = (key) => translations[lang.value][key] || key;

        const theme = computed(() => storeConfig.value.themeMode);
        const setTheme = (m) => { storeConfig.value.themeMode = m; document.documentElement.setAttribute('data-theme', m); };
        const toggleTheme = () => { const newTheme = storeConfig.value.themeMode === 'dark' ? 'light' : 'dark'; setTheme(newTheme); localStorage.setItem('theme', newTheme); };
        const toggleLang = () => { lang.value = lang.value === 'pt' ? 'en' : 'pt'; localStorage.setItem('lang', lang.value); };
        const initSettings = () => { const savedLang = localStorage.getItem('lang'); if (savedLang) lang.value = savedLang; };
        const showToast = (message, type = 'success') => { const id = Date.now(); toasts.value.push({ id, message, type }); setTimeout(() => { toasts.value = toasts.value.filter(t => t.id !== id); }, 3000); };

        const previewColors = computed(() => {
            const skinId = previewDesign.value;
            if (skinId === 'default') return { bg: '#00a896', text: '#ffffff', border: 'gold', bgSize: 'cover', bgPos: 'center' };
            if (skinId === 'custom') {
                return { 
                    bg: storeConfig.value.card_image_url ? `url(${storeConfig.value.card_image_url})` : (storeConfig.value.primary_color || '#00a896'), 
                    bgSize: storeConfig.value.card_image_url ? `${storeConfig.value.card_image_zoom || 100}%` : 'cover',
                    bgPos: storeConfig.value.card_image_url ? `calc(50% + ${storeConfig.value.card_image_pos_x || 0}px) calc(50% + ${storeConfig.value.card_image_pos_y || 0}px)` : 'center',
                    text: storeConfig.value.text_color || '#ffffff', border: storeConfig.value.border_color || '#ffffff'
                };
            }
            const skin = availableSkins.value.find(s => s.id === skinId);
            if (skin) {
                if (skin.image) return { bg: `url(${skin.image})`, text: '#ffffff', border: 'gold', bgSize: 'cover', bgPos: 'center' };
                if (skin.style) return { bg: skin.style.includes('background:') ? skin.style.split('background:')[1].split(';')[0] : '#333', text: '#ffffff', border: 'gold', bgSize: 'cover', bgPos: 'center' };
            }
            if (skinId === 'black') return { bg: '#1a1a1a', text: '#ffd166', border: '#ffd166', bgSize: 'cover', bgPos: 'center' };
            if (skinId === 'gold') return { bg: 'linear-gradient(45deg, #FFD700, #FDB931)', text: '#000000', border: 'rgba(0,0,0,0.5)', bgSize: 'cover', bgPos: 'center' };
            return { bg: '#00a896', text: '#ffffff', border: 'gold', bgSize: 'cover', bgPos: 'center' };
        });

        const filteredCards = computed(() => { 
            let result = cards.value;
            
            // 1. Filtro pelas Abas Inteligente
            if (activeTabScope.value !== 'all') {
                const mainScope = storeScopes.value.find(s => s.is_main);
                const isMainTab = mainScope && activeTabScope.value === mainScope.id;
                
                result = result.filter(c => {
                    // Se for a aba "Geral", puxa também os cartões antigos (sem scope_id)
                    if (isMainTab && (!c.scope_id || c.scope_id === '')) return true;
                    // Caso contrário, tem de bater certo
                    return c.scope_id === activeTabScope.value;
                });
            }
            
            // 2. Filtro pela Pesquisa
            if (searchQuery.value) {
                const q = searchQuery.value.toLowerCase(); 
                result = result.filter(c => c.customer_id.toLowerCase().includes(q) || c.last_name.toLowerCase().includes(q) || (c.email && c.email.toLowerCase().includes(q)) || (c.phone && c.phone.includes(q))); 
            }
            
            return result;
        });

        const fetchCards = async () => { try { const res = await fetch(api('/api/v1/admin/cards')); if (res.ok) cards.value = await res.json(); } catch(e) { console.error(e); } };
        const fetchAvailableSkins = async () => { try { const res = await fetch(api('/api/v1/system/skins')); if(res.ok) { const globals = await res.json(); const customSkin = { id: 'custom', name: 'Custom Brand', type: 'standard', style: `background: ${storeConfig.value.primary_color || '#00a896'}` }; availableSkins.value = [customSkin, ...globals]; } } catch(e) { console.error(e); } };

        const getPreviewStyle = (skin) => {
            if (!skin) return {};
            if(skin.id === 'default') return { background: '#00a896', color: '#fff', '--stamp-border': 'gold' };
            if(skin.id === 'black') return { background: '#1a1a1a', borderBottom: '3px solid #ffd166', color: '#fff', '--stamp-border': '#ffd166' };
            if(skin.id === 'custom') {
                const bg = storeConfig.value.card_image_url ? `url(${storeConfig.value.card_image_url})` : (storeConfig.value.primary_color || '#00a896');
                const bgSize = storeConfig.value.card_image_url ? `${storeConfig.value.card_image_zoom || 100}%` : 'cover';
                const bgPos = storeConfig.value.card_image_url ? `calc(50% + ${storeConfig.value.card_image_pos_x || 0}px) calc(50% + ${storeConfig.value.card_image_pos_y || 0}px)` : 'center';
                return { background: bg, backgroundSize: bgSize, backgroundPosition: bgPos, backgroundRepeat: 'no-repeat', color: storeConfig.value.text_color || '#fff', '--stamp-border': storeConfig.value.border_color || '#ffffff' };
            }
            if(skin.image) return { backgroundImage: `url(${skin.image})`, backgroundSize: 'cover', backgroundPosition: 'center', '--stamp-border': 'rgba(255,255,255,0.5)' };
            if(skin.style) return { background: skin.style.replace('background:', '').replace(';', ''), color: '#fff', '--stamp-border': 'rgba(255,255,255,0.5)' };
            return { background: '#ccc', '--stamp-border': 'rgba(0,0,0,0.2)' };
        };

        const changePage = (page) => { 
            if (currentPage.value === 'scanner' && page !== 'scanner') stopScanner();
            currentPage.value = page; 
            if (page === 'settings') settingsForm.value = { ...storeConfig.value };
            if (page === 'scanner') startScanner();
            if (page !== 'settings' && page !== 'skins' && page !== 'scanner') fetchCards(); 
        };

        const handleLogoUpload = (e) => { const r=new FileReader(); r.onload=(ev)=>{ settingsForm.value.logo_url=ev.target.result; }; r.readAsDataURL(e.target.files[0]); };
        const removeLogo = () => { settingsForm.value.logo_url=''; };

        const saveSettings = async () => { 
            await fetch(api('/api/v1/admin/settings'), { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(settingsForm.value) }); 
            showToast("Configuration Saved!", "success"); 
            storeConfig.value = { ...storeConfig.value, ...settingsForm.value };
            document.title = `Volto Store Admin | ${storeConfig.value.name}`;
        };

        const openCustomDesigner = () => {
            designForm.value = { background: storeConfig.value.primary_color || '#00a896', textColor: storeConfig.value.text_color || '#ffffff', borderColor: storeConfig.value.border_color || '#ffffff', image: storeConfig.value.card_image_url || null, zoom: storeConfig.value.card_image_zoom || 100, posX: storeConfig.value.card_image_pos_x || 0, posY: storeConfig.value.card_image_pos_y || 0 };
            showCustomDesigner.value = true;
        };

        const handleBgUpload = (e) => { const file = e.target.files[0]; if (!file) return; const reader = new FileReader(); reader.onload = (ev) => { designForm.value.image = ev.target.result; designForm.value.posX = 0; designForm.value.posY = 0; designForm.value.zoom = 100; }; reader.readAsDataURL(file); };
        const removeBgImage = () => { designForm.value.image = null; designForm.value.zoom = 100; designForm.value.posX = 0; designForm.value.posY = 0; };

        const saveCustomDesign = async () => {
            const newConfig = { primary_color: designForm.value.background, text_color: designForm.value.textColor, border_color: designForm.value.borderColor, card_image_url: designForm.value.image || '', card_image_zoom: parseInt(designForm.value.zoom) || 100, card_image_pos_x: parseInt(designForm.value.posX) || 0, card_image_pos_y: parseInt(designForm.value.posY) || 0 };
            storeConfig.value = { ...storeConfig.value, ...newConfig };
            await fetch(api('/api/v1/admin/settings'), { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(storeConfig.value) });
            showToast("Custom Brand Updated!", "success");
            showCustomDesigner.value = false;
            fetchAvailableSkins();
        };

        const updatePassword = async () => { const o=document.getElementById('pass-old').value, n=document.getElementById('pass-new').value; const res = await fetch(api('/api/v1/admin/update-password'), {method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({old:o, new:n})}); if(res.ok) showToast("Password Updated!", "success"); else showToast("Error updating password", "error"); };
        const updateGlobalSkin = async (s) => { await fetch(api('/api/v1/admin/update-skin'), {method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({design:s})}); activeSkin.value = s; showToast("Skin Activated!", "success"); };
        const openPreview = (d) => { previewDesign.value = d; showPreview.value = true; };
        const openEditModal = (c) => editingCard.value = { ...c };
        const saveEdit = async () => { await fetch(api('/api/v1/admin/update'), { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(editingCard.value) }); showToast("Customer Updated!", "success"); editingCard.value = null; fetchCards(); };
        const toggleConsent = async (c, type) => { const f = type === 'rgpd' ? 'rgpd_accepted' : 'marketing_accepted'; c[f] = !c[f]; await fetch(api('/api/v1/admin/update-consent'), { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ id: c.id, [f]: c[f] }) }); };
        const registerMember = async () => { 
            // Vai procurar qual é o Cartão Principal (Geral) para atribuir por defeito no registo manual
            const mainScope = storeScopes.value.find(s => s.is_main);

            const res = await fetch(api('/api/v1/cards'), { 
                method: 'POST', 
                headers: { 'Content-Type': 'application/json' }, 
                body: JSON.stringify({ 
                    customer_id: newMember.value.first_name, 
                    last_name: newMember.value.last_name, 
                    email: newMember.value.email, 
                    phone: newMember.value.phone, 
                    rgpd_accepted: newMember.value.rgpd, 
                    marketing_accepted: newMember.value.marketing,
                    // 👇 AGORA ENVIA O ÂMBITO 👇
                    scope_id: mainScope ? mainScope.id : '' 
                }) 
            }); 
            
            if(res.ok) { 
                showToast("Membro Registado!", "success"); 
                newMember.value = {first_name:'', last_name:'', email:'', phone:'', rgpd:false, marketing:false}; 
                changePage('customers'); 
            } else {
                showToast("Erro: Cliente já tem este cartão!", "error");
            }
        };
        const logout = async () => { await fetch(api('/api/v1/auth/logout'), { method: 'POST' }); window.location.href = "/"; };
        const calculateAvailable = (c) => Math.max(0, Math.floor(c.total_stamps / 10) - (c.total_redeemed_bonuses || 0));
        const addStampFromAdmin = async (c) => { await fetch(api(`/api/v1/cards/stamp?id=${c.id}`), {method:'POST'}); fetchCards(); showToast("Stamp Added", "success"); };
        const redeemFromAdmin = async (c) => { await fetch(api(`/api/v1/cards/use-reward?id=${c.id}`), {method:'POST'}); fetchCards(); showToast("Reward Redeemed", "success"); };
        const viewCard = (id) => window.open(`/card?store=${currentStore}&id=${id}`, '_blank');
        
        onMounted(() => { 
            initSettings(); 
            fetchCards(); 
            fetchScopes(); 
            fetch(api('/api/v1/system/config')).then(r => r.json()).then(d => { 
                storeConfig.value = { ...storeConfig.value, ...d }; 
                settingsForm.value = { ...storeConfig.value }; 
                activeSkin.value = d.card_skin || 'default';
                const localTheme = localStorage.getItem('theme');
                if (localTheme) { setTheme(localTheme); } else { setTheme(d.themeMode || 'dark'); }
                fetchAvailableSkins();
                document.title = `Volto Store Admin | ${d.name}`;
                if (d.status === 'expired') {
                    isAccountExpired.value = true;
                }
            }); 
        });

        const activeScopesList = computed(() => storeScopes.value.filter(s => s.is_active));

        return { 
            cards, currentPage, searchQuery, filteredCards, storeConfig, settingsForm,
            newMember, editingCard, lastScanStatus, activeSkin, availableSkins,
            showIconPicker, icons, selectIcon, saveSettings, handleLogoUpload, removeLogo,
            changePage, logout, registerMember, openEditModal, saveEdit, toggleConsent,
            calculateAvailable, addStampFromAdmin, redeemFromAdmin, viewCard, updatePassword, setTheme, toggleTheme,
            updateGlobalSkin, openPreview, showPreview, previewColors, getPreviewStyle,
            toasts, showToast, lang, toggleLang, t,
            showCustomDesigner, designForm, openCustomDesigner, handleBgUpload, removeBgImage, saveCustomDesign,
            theme, isDragging, startDrag, onDrag, stopDrag, handleMenuUpload, removeMenu,
            activeTabScope, deleteScope,
            storeScopes, activeScannerScope, predefinedScopeNames,
            newScopeName, newScopeIcon, createScope, toggleScope, activeScopesList,
            editingScopeId, editScopeTempName, editScopeTempIcon, startEditScope, cancelEditScope, saveEditScope, openIconPicker
        };
    }
};

Vue.createApp(AdminApp).mount('#app');