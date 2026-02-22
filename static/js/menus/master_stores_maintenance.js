const MasterStoresMaintenance = {
    template: `
    <div class="stores-view" style="max-width: 1200px; margin: 0 auto; padding-bottom: 50px;">
        
        <div class="header-bar" style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 40px;">
            <div class="page-title">
                <h2 style="margin: 0; font-size: 1.8rem; font-weight: 800; color: #111827;">Tenant Maintenance</h2>
                <p style="margin: 5px 0 0; color: #6b7280; font-size: 0.95rem;">Overview of {{ activeStores.length }} active SaaS partners.</p>
            </div>
            <div class="search-container" style="position: relative;">
                <span style="position: absolute; left: 12px; top: 50%; transform: translateY(-50%); color: #9ca3af;">🔍</span>
                <input v-model="search" placeholder="Search stores..." style="padding: 10px 10px 10px 38px; border: 1px solid #e5e7eb; border-radius: 8px; width: 250px; font-size: 0.9rem; outline: none;">
            </div>
        </div>
        
        <div style="font-size: 0.8rem; font-weight: 800; text-transform: uppercase; color: #111827; margin-bottom: 15px; letter-spacing: 0.5px;">ACTIVE PARTNERS</div>
        <div style="width: 100%; height: 1px; background: #e5e7eb; margin-bottom: 25px;"></div>

        <div class="grid-stores" style="display: grid; grid-template-columns: repeat(auto-fill, minmax(350px, 1fr)); gap: 25px; margin-bottom: 40px;">
            <div v-for="store in activeStores" :key="store.id" class="card-box" style="background: white; border: 1px solid #e5e7eb; border-radius: 12px; padding: 25px; box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1); position: relative; overflow: hidden;">
                
                <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 20px;">
                    <div style="display: flex; gap: 15px; align-items: center;">
                        <div style="width: 55px; height: 55px; background: #f9fafb; border-radius: 12px; display: flex; align-items: center; justify-content: center; font-size: 2rem; border: 1px solid #e5e7eb;">
                            <img v-if="store.logo_url" :src="store.logo_url" style="width: 100%; height: 100%; object-fit: contain;">
                            <span v-else>{{ store.stamp_icon || '🍳' }}</span>
                        </div>
                        <div>
                            <h3 style="margin: 0; font-size: 1.1rem; font-weight: 700; color: #111827;">{{ store.name }}</h3>
                            <span :style="getTierStyle(store.tier)" style="display: inline-block; padding: 3px 8px; border-radius: 6px; font-size: 0.65rem; font-weight: 800; text-transform: uppercase; margin-top: 5px;">{{ formatTier(store.tier) }}</span>
                        </div>
                    </div>
                    <span style="font-size: 0.6rem; font-weight: 800; color: #166534; background: #dcfce7; padding: 4px 10px; border-radius: 20px;">ACTIVE</span>
                </div>

                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 15px; background: #f9fafb; padding: 15px; border-radius: 10px; border: 1px solid #f3f4f6; margin-bottom: 15px;">
                    <div style="display: flex; flex-direction: column; gap: 5px;">
                        <span style="font-size: 0.65rem; font-weight: 700; color: #6b7280; text-transform: uppercase;">USAGE</span>
                        <div style="font-size: 0.85rem; font-weight: 600; color: #374151;">👥 {{ store.totalMembers || 0 }} Clients</div>
                        <div style="font-size: 0.85rem; font-weight: 600; color: #374151;">👔 1 / {{ store.max_users || 1 }} Staff</div>
                    </div>
                    <div style="display: flex; flex-direction: column; gap: 5px;">
                        <span style="font-size: 0.65rem; font-weight: 700; color: #6b7280; text-transform: uppercase;">HEALTH CHECK</span>
                        <div v-if="store.account_activated" style="font-size: 0.85rem; font-weight: 600; color: #059669;">✅ Onboarded</div>
                        <div v-else style="font-size: 0.85rem; font-weight: 600; color: #d97706;">⏳ Pending Login</div>
                        <div style="font-size: 0.8rem; color: #6b7280;">💳 {{ formatCycle(store.billing_cycle) }}</div>
                    </div>
                </div>

                <div :style="getCountdownStyle(store)" style="padding: 10px; border-radius: 8px; text-align: center; margin-bottom: 20px; font-weight: 700; font-size: 0.85rem; display: flex; justify-content: center; align-items: center; gap: 8px;">
                    <span>{{ getCountdownIcon(store) }}</span>
                    <span>{{ getCountdownText(store) }}</span>
                </div>

                <div style="display: flex; gap: 10px;">
                    <button @click="openEdit(store)" style="flex: 1; padding: 10px; border: 1px solid #e5e7eb; background: white; border-radius: 8px; font-weight: 600; font-size: 0.85rem; cursor: pointer; color: #374151;">⚙️ Config</button>
                    <a :href="'/?store=' + store.slug" target="_blank" style="flex: 1; padding: 10px; background: #2563eb; border: 1px solid #2563eb; border-radius: 8px; font-weight: 600; font-size: 0.85rem; color: white; text-align: center; text-decoration: none;">Dashboard ↗</a>
                </div>
            </div>
        </div>

        <div v-if="inactiveStores.length > 0">
            <div style="font-size: 0.8rem; font-weight: 800; text-transform: uppercase; color: #991b1b; margin-bottom: 15px; letter-spacing: 0.5px;">SUSPENDED / EXPIRED ACCOUNTS</div>
            <div style="width: 100%; height: 1px; background: #fecaca; margin-bottom: 25px;"></div>

            <div class="grid-stores" style="display: grid; grid-template-columns: repeat(auto-fill, minmax(350px, 1fr)); gap: 25px;">
                <div v-for="store in inactiveStores" :key="store.id" class="card-box" style="background: #fef2f2; border: 1px dashed #f87171; border-radius: 12px; padding: 25px; opacity: 0.9;">
                    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
                        <h3 style="margin: 0; color: #991b1b; font-size: 1.1rem;">{{ store.name }}</h3>
                        
                        <span v-if="!store.isActive" style="font-size: 0.6rem; font-weight: 800; color: #991b1b; background: #fee2e2; padding: 4px 10px; border-radius: 20px;">⛔ SUSPENDED</span>
                        <span v-else style="font-size: 0.6rem; font-weight: 800; color: #92400e; background: #fef3c7; padding: 4px 10px; border-radius: 20px;">⏳ EXPIRED</span>
                    </div>
                    
                    <p v-if="!store.isActive" style="font-size: 0.85rem; color: #7f1d1d; margin-bottom: 20px;">This store was manually suspended by the Master Admin.</p>
                    <p v-else style="font-size: 0.85rem; color: #7f1d1d; margin-bottom: 20px;">Trial or subscription expired {{ Math.abs(getDays(store.tier_expiration)) }} days ago.</p>
                    
                    <button @click="openEdit(store)" style="width: 100%; padding: 12px; background: #dc2626; color: white; border: none; border-radius: 8px; font-weight: 700; cursor: pointer;">MANAGE & REACTIVATE</button>
                </div>
            </div>
        </div>

        <div v-if="editingStore" style="position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.6); display: flex; align-items: center; justify-content: center; z-index: 1000; backdrop-filter: blur(4px);" @click.self="editingStore = null">
            <div class="modal" style="width: 550px; background: white; padding: 35px; border-radius: 16px; box-shadow: 0 25px 50px -12px rgba(0,0,0,0.25); max-height: 95vh; overflow-y: auto;">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 25px; border-bottom: 1px solid #e5e7eb; padding-bottom: 15px;">
                    <h2 style="margin: 0; font-size: 1.4rem; font-weight: 800; color: #111827;">SaaS Configuration</h2>
                    <span @click="editingStore = null" style="cursor: pointer; font-size: 1.5rem; color: #9ca3af;">✕</span>
                </div>
                
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin-bottom: 20px;">
                    <div>
                        <div class="form-group" style="margin-bottom: 15px;">
                            <label style="display: block; font-size: 0.7rem; font-weight: 800; margin-bottom: 6px; color: #6b7280; text-transform: uppercase;">STORE NAME</label>
                            <input v-model="editingStore.name" style="width: 100%; padding: 10px; border: 1px solid #d1d5db; border-radius: 6px; box-sizing: border-box;">
                        </div>
                        <div class="form-group" style="margin-bottom: 15px;">
                            <label style="display: block; font-size: 0.7rem; font-weight: 800; margin-bottom: 6px; color: #6b7280; text-transform: uppercase;">TIER PLAN</label>
                            <select v-model="editingStore.tier" style="width: 100%; padding: 10px; border: 1px solid #d1d5db; border-radius: 6px; box-sizing: border-box;">
                                <option value="free_trial">✨ Free Trial (30 Days)</option>
                                <option value="basic">⭐ Basic</option>
                                <option value="lite">🚀 Lite</option>
                                <option value="pro">👑 Pro</option>
                            </select>
                        </div>
                    </div>
                    <div>
                        <div class="form-group" style="margin-bottom: 15px;">
                            <label style="display: block; font-size: 0.7rem; font-weight: 800; margin-bottom: 6px; color: #6b7280; text-transform: uppercase;">BILLING CYCLE</label>
                            <select v-model="editingStore.billing_cycle" style="width: 100%; padding: 10px; border: 1px solid #d1d5db; border-radius: 6px; box-sizing: border-box;">
                                <option value="monthly">Monthly</option>
                                <option value="quarterly">Quarterly</option>
                                <option value="biannual">Biannual (6M)</option>
                                <option value="annual">Annual (1Y)</option>
                            </select>
                        </div>
                        <div class="form-group" style="margin-bottom: 15px;">
                            <label style="display: block; font-size: 0.7rem; font-weight: 800; margin-bottom: 6px; color: #6b7280; text-transform: uppercase;">NEXT PAYMENT / EXPIRY</label>
                            <input type="date" v-model="editingStore.tier_expiration" style="width: 100%; padding: 10px; border: 1px solid #d1d5db; border-radius: 6px; box-sizing: border-box;">
                        </div>
                    </div>
                </div>

                <div style="background: #f3f4f6; padding: 15px; border-radius: 8px; margin-bottom: 25px; border: 1px solid #e5e7eb;">
                    <div style="font-size: 0.75rem; font-weight: 800; color: #374151; margin-bottom: 10px; text-transform: uppercase;">ADMIN ACCESS</div>
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                        <span style="font-size: 0.85rem; color: #4b5563;">{{ editingStore.admin_email }}</span>
                        <button @click="emit('show-toast', 'Password Reset Link Sent via Email')" style="border: 1px solid #d1d5db; background: white; padding: 6px 12px; border-radius: 6px; font-weight: 600; font-size: 0.75rem; cursor: pointer; color: #2563eb;">📧 Reset Password</button>
                    </div>
                </div>

                <button @click="saveChanges" style="width: 100%; padding: 14px; background: #2563eb; color: white; border: none; border-radius: 8px; font-weight: 700; font-size: 0.95rem; cursor: pointer; box-shadow: 0 4px 6px -1px rgba(37, 99, 235, 0.2); margin-bottom: 15px;">SAVE CONFIGURATION</button>

                <div style="border-top: 1px solid #e5e7eb; padding-top: 20px; display:flex; justify-content: space-between; align-items: center;">
                    <div>
                        <div style="font-size: 0.8rem; font-weight: 800; color: #991b1b;">STORE STATUS</div>
                        <div style="font-size: 0.75rem; color: #6b7280;">Suspend access manually?</div>
                    </div>
                    <button @click="toggleStatus(editingStore)" :style="editingStore.isActive ? 'background: #fee2e2; color: #991b1b; border: 1px solid #fecaca;' : 'background: #dcfce7; color: #166534; border: 1px solid #bbf7d0;'" style="padding: 8px 16px; border-radius: 6px; font-weight: 700; cursor: pointer; font-size: 0.8rem;">
                        {{ editingStore.isActive ? '⛔ SUSPEND STORE' : '✅ REACTIVATE STORE' }}
                    </button>
                </div>
            </div>
        </div>
    </div>
    `,
    setup(props, { emit }) {
        const { ref, computed, onMounted } = Vue;
        const stores = ref([]);
        const search = ref('');
        const editingStore = ref(null);

        const fetchStores = async () => { 
            try { 
                const res = await fetch('/api/v1/master/stores'); 
                if(res.ok) stores.value = await res.json() || []; 
            } catch(e) { console.error(e); }
        };

        // Função utilitária para calcular os dias
        const getDays = (dateStr) => {
            if(!dateStr) return 0;
            const target = new Date(dateStr);
            const now = new Date();
            const diff = target - now;
            return Math.ceil(diff / (1000 * 60 * 60 * 24));
        };

        // Identifica automaticamente se a conta expirou
        const isStoreExpired = (s) => getDays(s.tier_expiration) <= 0;

        const filtered = computed(() => { 
            if(!stores.value) return []; 
            return stores.value.filter(s => s.name.toLowerCase().includes(search.value.toLowerCase())); 
        });
        
        // --- A MAGIA DA ARRUMAÇÃO INTELIGENTE ---
        // Lojas ativas E que ainda têm dias restantes
        const activeStores = computed(() => filtered.value.filter(s => s.isActive && !isStoreExpired(s)));
        
        // Lojas suspensas manualmente OU que já expiraram
        const inactiveStores = computed(() => filtered.value.filter(s => !s.isActive || isStoreExpired(s)));

        const toggleStatus = async (s) => {
            const newStatus = !s.isActive;
            const confirmMsg = newStatus ? "Reactivate store?" : "Suspend this store? Users will lose access immediately.";
            if(!confirm(confirmMsg)) return;

            const res = await fetch('/api/v1/master/stores/status', { 
                method: 'POST', 
                headers: {'Content-Type': 'application/json'}, 
                body: JSON.stringify({id: s.id, isActive: newStatus}) 
            });
            if(res.ok) { 
                fetchStores(); 
                if(editingStore.value && editingStore.value.id === s.id) {
                    editingStore.value.isActive = newStatus;
                }
                emit('show-toast', newStatus ? "Store Reactivated" : "Store Suspended"); 
            }
        };

        const saveChanges = async () => { 
            const payload = { ...editingStore.value };
            if (payload.tier_expiration) {
                payload.tier_expiration = new Date(payload.tier_expiration).toISOString();
            }

            const res = await fetch('/api/v1/master/stores/update', { 
                method: 'POST', 
                headers: {'Content-Type': 'application/json'}, 
                body: JSON.stringify(payload) 
            }); 
            if(res.ok) { 
                editingStore.value = null; 
                fetchStores(); 
                emit('show-toast', "Configuration Saved"); 
            } else {
                emit('show-toast', "Error saving changes", "error");
            }
        };

        const handleLogoUpload = (e) => { 
            const file = e.target.files[0]; 
            const reader = new FileReader(); 
            reader.onload = (ev) => editingStore.value.logo_url = ev.target.result; 
            reader.readAsDataURL(file); 
        };

        const formatTier = (tier) => {
            const map = { 'free_trial': 'Free Trial', 'basic': 'Basic', 'lite': 'Lite', 'pro': 'Pro' };
            return map[tier] || tier;
        };

        const getTierStyle = (tier) => {
            const map = {
                'free_trial': 'background: #fff7ed; color: #c2410c; border: 1px solid #ffedd5;', 
                'basic': 'background: #f3f4f6; color: #374151; border: 1px solid #e5e7eb;', 
                'lite': 'background: #eff6ff; color: #1d4ed8; border: 1px solid #dbeafe;', 
                'pro': 'background: #fdf2f8; color: #be185d; border: 1px solid #fce7f3;' 
            };
            return map[tier] || map['basic'];
        };

        const formatCycle = (cycle) => {
            const map = { 'monthly': 'Monthly Billing', 'quarterly': 'Quarterly Billing', 'biannual': 'Biannual Billing', 'annual': 'Annual Billing' };
            return map[cycle] || 'Monthly Billing';
        };

        const getCountdownText = (s) => {
            const days = getDays(s.tier_expiration);
            if(s.tier === 'free_trial') {
                return days > 0 ? `${days} days left in trial` : 'Trial Expired';
            }
            return days > 0 ? `Next payment in ${days} days` : 'Payment Overdue';
        };

        const getCountdownIcon = (s) => {
            const days = getDays(s.tier_expiration);
            if (days <= 0) return '🚨'; 
            if (days < 5) return '⚠️'; 
            return '🕒'; 
        };

        const getCountdownStyle = (s) => {
            const days = getDays(s.tier_expiration);
            if (days <= 0) return 'background: #fee2e2; color: #991b1b; border: 1px solid #fecaca;';
            if (days <= 5) return 'background: #fef3c7; color: #92400e; border: 1px solid #fde68a;';
            if (s.tier === 'free_trial') return 'background: #ecfdf5; color: #047857; border: 1px solid #a7f3d0;';
            return 'background: #f3f4f6; color: #4b5563; border: 1px solid #e5e7eb;';
        };

        const openEdit = (s) => {
            let dateStr = '';
            if (s.tier_expiration) {
                dateStr = s.tier_expiration.split('T')[0];
            }
            editingStore.value = { ...s, tier_expiration: dateStr };
        };

        Vue.onMounted(fetchStores);

        return { 
            stores, search, activeStores, inactiveStores, editingStore, toggleStatus, openEdit, saveChanges, handleLogoUpload, emit,
            formatTier, getTierStyle, formatCycle, getCountdownText, getCountdownStyle, getCountdownIcon, getDays
        };
    }
};