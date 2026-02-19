const { createApp, ref, computed, onMounted } = Vue;

createApp({
    setup() {
        const urlParams = new URLSearchParams(window.location.search);
        const currentStore = urlParams.get('store');
        const api = (url) => currentStore ? `${url}${url.includes('?') ? '&' : '?'}store=${currentStore}` : url;

        const card = ref({ id: null, total_stamps: 0, stamps_count: 0 });
        const isFlipped = ref(false);
        const cardId = urlParams.get('id');
        
        // Todas as novas variáveis de design incluídas
        const storeConfig = ref({ 
            name: 'Store', 
            logo_url: '', 
            themeMode: 'dark', 
            primary_color: '#00a896',
            text_color: '#ffffff',
            border_color: '#ffffff',
            card_image_url: '',
            card_image_zoom: 100,
            card_image_pos_x: 0,
            card_image_pos_y: 0,
            card_skin: 'default',
            bronzeThreshold: 15, silverThreshold: 40, goldThreshold: 100, 
            stamp_icon: '🍳' 
        });

        const activeSkinData = ref(null);

        const customerTier = computed(() => {
            const total = card.value.total_stamps || 0;
            const b = storeConfig.value.bronzeThreshold || 15, s = storeConfig.value.silverThreshold || 40, g = storeConfig.value.goldThreshold || 100;
            if (total >= g) return { name: 'Gold', color: '#ffd166', glow: 'rgba(255, 209, 102, 0.6)' };
            if (total >= s) return { name: 'Silver', color: '#bdc3c7', glow: 'rgba(189, 195, 199, 0.6)' };
            if (total >= b) return { name: 'Bronze', color: '#cd7f32', glow: 'rgba(205, 127, 50, 0.6)' };
            return { name: '', color: '', glow: 'transparent' };
        });

        const tierGlowStyle = computed(() => ({ '--tier-glow': customerTier.value.glow }));
        
        const applyTheme = () => { document.documentElement.style.setProperty('--page-bg', storeConfig.value.themeMode === 'light' ? '#f0f2f5' : '#1a1a1a'); };
        
        const fetchSettings = async () => {
            const res = await fetch(api('/api/v1/system/config'));
            if (res.ok) { 
                storeConfig.value = { ...storeConfig.value, ...(await res.json()) }; 
                applyTheme(); 
            }
        };

        const fetchSkinDetails = async () => {
            try {
                const res = await fetch(api('/api/v1/system/skins'));
                if (res.ok) {
                    const skins = await res.json();
                    activeSkinData.value = skins.find(s => s.id === storeConfig.value.card_skin);
                }
            } catch(e) { console.error("Error fetching skins", e); }
        };

        // LÓGICA DE CORES & IMAGEM (Idêntica ao Theme Designer do Admin)
        const themeStyles = computed(() => {
            const skinId = storeConfig.value.card_skin || 'default';
            
            // Variáveis por defeito
            let styles = { 
                bg: '#00a896', bgImage: 'none', bgSize: 'cover', bgPos: 'center', 
                color: '#ffffff', stampBorder: 'gold' 
            };

            // 1. Custom Brand (Aplica tudo o que configuraste no designer)
            if (skinId === 'custom') {
                styles.bg = storeConfig.value.card_image_url ? 'transparent' : (storeConfig.value.primary_color || '#00a896');
                styles.bgImage = storeConfig.value.card_image_url ? `url(${storeConfig.value.card_image_url})` : 'none';
                styles.bgSize = storeConfig.value.card_image_url ? `${storeConfig.value.card_image_zoom || 100}%` : 'cover';
                styles.bgPos = storeConfig.value.card_image_url ? `calc(50% + ${storeConfig.value.card_image_pos_x || 0}px) calc(50% + ${storeConfig.value.card_image_pos_y || 0}px)` : 'center';
                styles.color = storeConfig.value.text_color || '#ffffff';
                styles.stampBorder = storeConfig.value.border_color || '#ffffff';
                return styles;
            } 
            
            // 2. Skins Globais da BD
            if (activeSkinData.value) {
                if(activeSkinData.value.image) {
                    styles.bgImage = `url(${activeSkinData.value.image})`;
                    return styles;
                }
                if(activeSkinData.value.style) {
                    styles.bg = activeSkinData.value.style.replace('background:', '').replace(';', '').trim();
                    return styles;
                }
            }

            // 3. Fallbacks Hardcoded (Segurança)
            if (skinId === 'black') {
                styles.bg = '#1a1a1a'; styles.color = '#ffd166'; styles.stampBorder = '#ffd166';
            }
            if (skinId === 'gold') {
                styles.bg = 'linear-gradient(45deg, #FFD700, #FDB931)'; styles.color = '#000000'; styles.stampBorder = 'rgba(0,0,0,0.5)';
            }
            
            return styles;
        });

        const availableRewards = computed(() => Math.max(0, Math.floor(card.value.total_stamps / 10) - (card.value.total_redeemed_bonuses || 0)));
        const fetchCard = async () => { const res = await fetch(api(`/api/v1/cards/status?id=${cardId}`)); if (res.ok) card.value = await res.json(); };
        const qrCodeUrl = computed(() => api('/api/v1/qrcode?id=' + card.value.id));

        onMounted(async () => { 
            fetchCard(); 
            await fetchSettings();
            fetchSkinDetails(); 
        });
        
        return { card, isFlipped, availableRewards, themeStyles, storeConfig, customerTier, tierGlowStyle, qrCodeUrl };
    }
}).mount('#app');