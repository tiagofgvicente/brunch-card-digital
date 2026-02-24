const { createApp, ref, computed, onMounted, onUnmounted } = Vue;

createApp({
    setup() {
        const urlParams = new URLSearchParams(window.location.search);
        const currentStore = urlParams.get('store');
        
        const hasStore = ref(!!currentStore && currentStore !== 'null');
        const api = (url) => hasStore.value ? `${url}${url.includes('?') ? '&' : '?'}store=${currentStore}` : url;

        // O is_verified começa a true para não assustar enquanto carrega
        const card = ref({ id: null, customer_id: '', last_name: '', email: '', phone: '', total_stamps: 0, stamps_count: 0, is_verified: true });
        const myWalletCards = ref([]); 
        
        const notifications = ref([]);
        const unreadCount = computed(() => notifications.value.filter(n => !n.is_read).length);

        const isFlipped = ref(false);
        const cardId = urlParams.get('id');
        
        const toasts = ref([]);
        const showToast = (message, type = 'success') => { 
            const id = Date.now(); toasts.value.push({ id, message, type }); 
            setTimeout(() => { toasts.value = toasts.value.filter(t => t.id !== id); }, 3000); 
        };

        const isMenuOpen = ref(false);
        const showRedeemModal = ref(false);
        const currentView = ref(hasStore.value ? 'card' : 'my_cards'); 
        
        const isChangingCard = ref(false);

        const toggleMenu = () => { isMenuOpen.value = !isMenuOpen.value; };
        
        const changeView = (view) => { 
            currentView.value = view; isMenuOpen.value = false; isEditingProfile.value = false; 
            if (view === 'my_cards' && card.value.email) fetchWalletCards(card.value.email);
            
            if (view === 'notifications' && card.value.email && unreadCount.value > 0) {
                notifications.value.forEach(n => n.is_read = true); 
                fetch('/api/v1/public/wallet-notifications/read', {
                    method: 'POST', headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ email: card.value.email })
                });
            }
        };
        
        const logout = () => { window.location.href = "/"; };

        const isEditingProfile = ref(false);
        const profileForm = ref({});

        const startEditProfile = () => { profileForm.value = { ...card.value }; isEditingProfile.value = true; };

        const saveProfile = async () => {
            try {
                const res = await fetch(hasStore.value ? api('/api/v1/admin/update') : '/api/v1/public/wallet-update', {
                    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(profileForm.value)
                });
                if (res.ok) {
                    card.value = { ...card.value, ...profileForm.value };
                    isEditingProfile.value = false; showToast("Perfil atualizado com sucesso!");
                } else showToast("Erro ao atualizar o perfil.", "error");
            } catch (e) { showToast("Erro de ligação.", "error"); }
        };

        const globalQrUrl = computed(() => {
            const data = JSON.stringify({ action: 'global_register', email: card.value.email, first: card.value.customer_id, last: card.value.last_name, phone: card.value.phone });
            return `https://api.qrserver.com/v1/create-qr-code/?size=250x250&data=${encodeURIComponent(data)}`;
        });

        const defaultAppColor = hasStore.value ? '#00a896' : '#2563eb';

        const storeConfig = ref({ 
            name: hasStore.value ? 'Store' : 'Volto Wallet', logo_url: '', themeMode: 'dark', 
            primary_color: defaultAppColor, text_color: '#ffffff', border_color: '#ffffff', 
            card_image_url: '', card_image_zoom: 100, card_image_pos_x: 0, card_image_pos_y: 0, card_scope: 'Geral', 
            social_instagram: '', social_facebook: '', social_twitter: '', social_whatsapp: '', social_tiktok: '', social_youtube: '', social_website: '', 
            menu_url: '', location_url: '', card_skin: 'default', 
            bronzeThreshold: 15, silverThreshold: 40, goldThreshold: 100, stamp_icon: hasStore.value ? '🍳' : '💳' 
        });

        const activeSkinData = ref(null);

        const customerTier = computed(() => {
            const total = card.value.total_stamps || 0;
            const b = storeConfig.value.bronzeThreshold || 15, s = storeConfig.value.silverThreshold || 40, g = storeConfig.value.goldThreshold || 100;
            if (total >= g) return { name: 'Gold', color: '#ffd166', glow: 'rgba(255, 209, 102, 0.6)' };
            if (total >= s) return { name: 'Silver', color: '#bdc3c7', glow: 'rgba(189, 195, 199, 0.6)' };
            if (total >= b) return { name: 'Bronze', color: '#cd7f32', glow: 'rgba(205, 127, 50, 0.6)' };
            return { name: 'New Member', color: 'inherit', glow: 'transparent' };
        });

        const tierGlowStyle = computed(() => ({ '--tier-glow': customerTier.value.glow }));
        
        const applyTheme = (overrideColor) => { 
            const isDark = storeConfig.value.themeMode !== 'light';
            const baseBg = isDark ? '#1a1a1a' : '#f0f2f5';
            const textColor = isDark ? '#ffffff' : '#1a1a1a';
            const pColor = overrideColor || storeConfig.value.primary_color || defaultAppColor;
            
            document.documentElement.style.setProperty('--page-bg', baseBg);
            document.documentElement.style.setProperty('--text-main', textColor);
            document.documentElement.style.setProperty('--page-grad', `radial-gradient(circle at 50% -10%, ${pColor}40 0%, transparent 80%)`);
            document.documentElement.style.setProperty('--primary', pColor);
        };
        
        const fetchSettings = async (specificStore = null) => {
            const endpoint = specificStore ? `/api/v1/system/config?store=${specificStore}` : api('/api/v1/system/config');
            const res = await fetch(endpoint);
            if (res.ok) {
                const data = await res.json();
                const socialDefaults = {
                    social_instagram: '', social_facebook: '', social_twitter: '', social_whatsapp: '', 
                    social_tiktok: '', social_youtube: '', social_website: '', 
                    menu_url: '', location_url: ''
                };
                storeConfig.value = { ...storeConfig.value, ...socialDefaults, ...data }; 
                applyTheme(); 
            }
        };

        const fetchSkinDetails = async (targetStore = null) => {
            try {
                const endpoint = targetStore ? `/api/v1/system/skins?store=${targetStore}` : api('/api/v1/system/skins');
                const res = await fetch(endpoint);
                if (res.ok) {
                    const skins = await res.json();
                    activeSkinData.value = skins.find(s => s.id === storeConfig.value.card_skin);
                }
            } catch(e) {}
        };

        const fetchWalletCards = async (email) => {
            if(!email) return;
            try {
                const res = await fetch(`/api/v1/public/wallet-mycards?email=${encodeURIComponent(email)}`);
                if(res.ok) myWalletCards.value = await res.json();
            } catch(e) {}
        };

        const fetchNotifications = async (email) => {
            if(!email) return;
            try {
                const res = await fetch(`/api/v1/public/wallet-notifications?email=${encodeURIComponent(email)}`);
                if(res.ok) notifications.value = await res.json();
            } catch(e) {}
        };

        // 👇 FETCH CARD ATUALIZADO (A MÁGICA ESTÁ AQUI) 👇
        const fetchCard = async () => { 
            const endpoint = hasStore.value ? `/api/v1/cards/status?id=${cardId}&store=${currentStore}` : `/api/v1/public/wallet-profile?id=${cardId}`;
            const res = await fetch(endpoint); 
            if (res.ok) {
                const data = await res.json(); 
                
                // Mapeia inteligentemente quer venha de uma loja ou da wallet
                card.value = { 
                    ...card.value, 
                    ...data,
                    customer_id: data.customer_id || data.first_name || '', // Unifica o nome
                    is_verified: data.is_verified !== undefined ? data.is_verified : true // Protege contra false hiding
                }; 
                
                if(!hasStore.value && card.value.email) {
                    fetchWalletCards(card.value.email);
                    fetchNotifications(card.value.email); 
                }
            }
        };

        const calcRewards = (stamps, redeemed) => Math.max(0, Math.floor(stamps / 10) - (redeemed || 0));
        const availableRewards = computed(() => calcRewards(card.value.total_stamps, card.value.total_redeemed_bonuses));

        const activeWalletStoreSlug = ref(null);
        const activeWalletCardId = ref(null);
        
        const currentIndex = computed(() => {
            if (!activeWalletCardId.value || myWalletCards.value.length === 0) return -1;
            return myWalletCards.value.findIndex(c => c.id === activeWalletCardId.value);
        });

        const nextWalletCard = () => {
            if (myWalletCards.value.length <= 1 || isChangingCard.value) return;
            let nextIdx = currentIndex.value + 1;
            if (nextIdx >= myWalletCards.value.length) nextIdx = 0; 
            const nextCard = myWalletCards.value[nextIdx];
            openWalletCard(nextCard.store_slug, nextCard.id);
        };

        const prevWalletCard = () => {
            if (myWalletCards.value.length <= 1 || isChangingCard.value) return;
            let prevIdx = currentIndex.value - 1;
            if (prevIdx < 0) prevIdx = myWalletCards.value.length - 1; 
            const prevCard = myWalletCards.value[prevIdx];
            openWalletCard(prevCard.store_slug, prevCard.id);
        };

        const openWalletCard = async (storeSlug, cId) => { 
            if (isChangingCard.value) return; 
            const isFromList = currentView.value === 'my_cards';
            if (!isFromList) isChangingCard.value = true;
            
            const [configRes, cardRes] = await Promise.all([
                fetch(`/api/v1/system/config?store=${storeSlug}`),
                fetch(`/api/v1/cards/status?id=${cId}&store=${storeSlug}`)
            ]);

            const newConfig = configRes.ok ? await configRes.json() : null;
            const newCard = cardRes.ok ? await cardRes.json() : null;

            const finalizeSwap = async () => {
                isFlipped.value = false; 
                activeWalletCardId.value = cId;
                activeWalletStoreSlug.value = storeSlug;
                
                if (newConfig) { 
                    const socialDefaults = { social_instagram: '', social_facebook: '', social_twitter: '', social_whatsapp: '', social_tiktok: '', social_youtube: '', social_website: '', menu_url: '', location_url: '' };
                    storeConfig.value = { ...storeConfig.value, ...socialDefaults, ...newConfig }; 
                    applyTheme(); 
                    await fetchSkinDetails(storeSlug); 
                }
                if (newCard) { card.value = newCard; currentView.value = 'wallet_active_card'; }
                isChangingCard.value = false; 
            };

            if (isFromList) finalizeSwap();
            else setTimeout(finalizeSwap, 250);
        };

        const backToWallet = async () => {
            if (isChangingCard.value) return;
            isChangingCard.value = true;
            activeWalletCardId.value = null;
            activeWalletStoreSlug.value = null;
            activeSkinData.value = null; 
            await fetchCard();
            
            setTimeout(() => {
                storeConfig.value.name = 'Volto Wallet';
                storeConfig.value.logo_url = '';
                storeConfig.value.card_skin = 'default';
                applyTheme('#2563eb'); 
                currentView.value = 'my_cards';
                isChangingCard.value = false;
            }, 250);
        };

        const handleKeydown = (e) => {
            if (currentView.value === 'wallet_active_card' && myWalletCards.value.length > 1 && !isMenuOpen.value && !showRedeemModal.value && !isChangingCard.value) {
                if (e.key === 'ArrowRight') nextWalletCard();
                if (e.key === 'ArrowLeft') prevWalletCard();
                if (e.key === 'Escape') backToWallet();
            }
        };

        const themeStyles = computed(() => {
            if (card.value.scope_is_active === false) {
                return { bg: '#333333', bgImage: 'none', bgSize: 'cover', bgPos: 'center', color: 'rgba(255,255,255,0.5)', stampBorder: 'rgba(255,255,255,0.2)' };
            }

            const skinId = storeConfig.value.card_skin || 'default';
            const mainColor = storeConfig.value.primary_color || defaultAppColor;
            let styles = { bg: mainColor, bgImage: 'none', bgSize: 'cover', bgPos: 'center', color: '#ffffff', stampBorder: 'gold' };
            
            if (skinId === 'custom') {
                styles.bg = storeConfig.value.card_image_url ? 'transparent' : mainColor;
                styles.bgImage = storeConfig.value.card_image_url ? `url(${storeConfig.value.card_image_url})` : 'none';
                styles.bgSize = storeConfig.value.card_image_url ? `${storeConfig.value.card_image_zoom || 100}%` : 'cover';
                styles.bgPos = storeConfig.value.card_image_url ? `calc(50% + ${storeConfig.value.card_image_pos_x || 0}px) calc(50% + ${storeConfig.value.card_image_pos_y || 0}px)` : 'center';
                styles.color = storeConfig.value.text_color || '#ffffff';
                styles.stampBorder = storeConfig.value.border_color || '#ffffff';
                return styles;
            } 
            
            if (activeSkinData.value) {
                if(activeSkinData.value.image) { 
                    styles.bgImage = `url(${activeSkinData.value.image})`; 
                } else if (activeSkinData.value.colorBg) {
                    styles.bg = activeSkinData.value.colorBg;
                }
                styles.color = activeSkinData.value.colorText || '#ffffff';
                styles.stampBorder = activeSkinData.value.colorBorder || 'gold';
                return styles;
            }
            return styles;
        });

        // 👇 PROTEÇÃO CONTRA O ERRO 502 DE BAD GATEWAY 👇
        const qrCodeUrl = computed(() => {
            const targetId = hasStore.value ? card.value.id : (activeWalletCardId.value || card.value.id);
            const targetStore = hasStore.value ? currentStore : activeWalletStoreSlug.value;
            if (!targetId || !targetStore) return ''; // Proteção adicionada
            const url = `${window.location.origin}/card?store=${targetStore}&id=${targetId}`;
            return `https://api.qrserver.com/v1/create-qr-code/?size=250x250&data=${encodeURIComponent(url)}`;
        });

        const redeemQrUrl = computed(() => {
            const targetId = hasStore.value ? card.value.id : activeWalletCardId.value;
            const targetStore = hasStore.value ? currentStore : activeWalletStoreSlug.value;
            if (!targetId || !targetStore) return ''; // Proteção adicionada
            const data = JSON.stringify({ action: 'redeem', id: targetId, store: targetStore });
            return `https://api.qrserver.com/v1/create-qr-code/?size=250x250&data=${encodeURIComponent(data)}`;
        });

        const formatDate = (dateStr) => {
            const d = new Date(dateStr);
            return d.toLocaleDateString('pt-PT', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' });
        };

        onMounted(async () => { 
            window.addEventListener('keydown', handleKeydown); 
            if (window.location.hash === '#global_qr') currentView.value = 'global_qr';
            await fetchCard(); 
            if (hasStore.value) { 
                await fetchSettings(); 
                fetchSkinDetails(); 
            } 
            else { applyTheme('#2563eb'); }
        });

        onUnmounted(() => { window.removeEventListener('keydown', handleKeydown); });
        
        return { 
            card, isFlipped, calcRewards, availableRewards, themeStyles, storeConfig, customerTier, tierGlowStyle, qrCodeUrl, redeemQrUrl,
            isMenuOpen, currentView, toggleMenu, changeView, logout, globalQrUrl,
            isEditingProfile, profileForm, startEditProfile, saveProfile, toasts, hasStore, myWalletCards, openWalletCard, backToWallet,
            showRedeemModal, nextWalletCard, prevWalletCard, isChangingCard,
            notifications, unreadCount, formatDate 
        };
    }
}).mount('#app');