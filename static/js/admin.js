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
        const lang = ref('pt'); // Default (pt / en)
        
        const showIconPicker = ref(false);
        const icons = ['🍳','🍔','🍕','🌭','🥪','🌮','🥗','🥐','🥯','🥞','☕','🍺','🍷','🍹','🥤','🧋','🍩','🍪','🍰','🍦','✂️','💇','💅','💄','💈','🏋️','🧘','🚗','🚲','🎮','🐾','🐶','🐱','📚','💊','🦷','🕶️','💍','👠','🧢'];

        const storeConfig = ref({ 
            name: 'Store', logo_url: '', themeMode: 'dark', 
            primary_color: '#00a896', stamp_icon: '🍳',
            bronzeThreshold: 15, silverThreshold: 40, goldThreshold: 100 
        });
        
        const newMember = ref({ first_name: '', last_name: '', email: '', phone: '', rgpd: false, marketing: false });
        const editingCard = ref(null);
        const lastScanStatus = ref('');
        const scannerInput = ref(null);
        const availableSkins = ref([]);

        // --- COMPUTED ---
        const previewColors = computed(() => {
            const skinId = previewDesign.value;
            if (skinId === 'custom') return { bg: storeConfig.value.primary_color || '#00a896', text: '#ffffff' };
            
            const skin = availableSkins.value.find(s => s.id === skinId);
            if (skin) {
                if (skin.image) return { bg: `url(${skin.image})`, text: '#ffffff' };
                if (skin.style) {
                     const bg = skin.style.includes('background:') ? skin.style.split('background:')[1].split(';')[0] : '#333';
                     return { bg: bg, text: '#ffffff' };
                }
            }
            if (skinId === 'black') return { bg: '#1a1a1a', text: '#ffd166' };
            if (skinId === 'gold') return { bg: 'linear-gradient(45deg, #FFD700, #FDB931)', text: '#000000' };
            return { bg: storeConfig.value.primary_color || '#00a896', text: '#ffffff' };
        });

        const filteredCards = computed(() => {
            if (!searchQuery.value) return cards.value;
            const q = searchQuery.value.toLowerCase();
            return cards.value.filter(c => c.customer_id.toLowerCase().includes(q) || c.last_name.toLowerCase().includes(q) || (c.email && c.email.toLowerCase().includes(q)) || (c.phone && c.phone.includes(q)));
        });

        // --- ACTIONS ---
        const showToast = (message, type = 'success') => {
            const id = Date.now();
            toasts.value.push({ id, message, type });
            setTimeout(() => { toasts.value = toasts.value.filter(t => t.id !== id); }, 3000);
        };

        const toggleLang = () => {
            lang.value = lang.value === 'pt' ? 'en' : 'pt';
            localStorage.setItem('admin_lang', lang.value);
            // Aqui podes adicionar lógica futura para traduzir a interface
        };

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

        const handleLogoUpload = (e) => { const r=new FileReader(); r.onload=(ev)=>{ storeConfig.value.logo_url=ev.target.result; }; r.readAsDataURL(e.target.files[0]); };
        const removeLogo = () => { storeConfig.value.logo_url=''; };

        const updatePassword = async () => { 
            const o=document.getElementById('pass-old').value, n=document.getElementById('pass-new').value; 
            const res = await fetch(api('/api/v1/admin/update-password'), {method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({old:o, new:n})}); 
            if(res.ok) showToast("Password Updated!", "success"); else showToast("Error updating password", "error");
        };

        const setTheme = (m) => { 
            storeConfig.value.themeMode = m; 
            // Injeta o atributo no HTML para o CSS funcionar
            document.documentElement.setAttribute('data-theme', m);
        };
        const toggleTheme = () => {
            const newTheme = storeConfig.value.themeMode === 'dark' ? 'light' : 'dark';
            setTheme(newTheme);
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
            fetchCards(); 
            lang.value = localStorage.getItem('admin_lang') || 'pt';
            fetch(api('/api/v1/system/config')).then(r => r.json()).then(d => { 
                storeConfig.value = { ...storeConfig.value, ...d }; 
                activeSkin.value = d.card_skin || 'default';
                setTheme(d.themeMode || 'dark');
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
            toasts, showToast, lang, toggleLang
        };
    }
};

Vue.createApp(AdminApp).mount('#app');