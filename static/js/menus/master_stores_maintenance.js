const MasterStoresMaintenance = {
    template: `
    <div class="stores-view">
        <div class="header-bar">
            <div class="page-title"><h2>Tenant Maintenance</h2><p>Managing {{ activeStores.length }} active SaaS partners.</p></div>
            <div class="search-container"><span class="search-icon">🔍</span><input v-model="search" class="search-input" placeholder="Search stores..."></div>
        </div>
        
        <div class="section-title">Active Partners</div>
        <div class="grid-stores">
            <div v-for="store in activeStores" :key="store.id" class="card-box">
                <div class="store-header">
                    <div class="store-branding-row">
                        <div class="store-icon"><img v-if="store.logo_url" :src="store.logo_url"><span v-else>{{ store.stamp_icon || '🍳' }}</span></div>
                        <div class="store-info"><h3>{{ store.name }}</h3><span :class="['tier-badge', 'tier-' + (store.tier || 'free_trial')]">{{ store.tier }}</span></div>
                    </div>
                    <span class="status-badge status-active" @click="toggleStatus(store)">ACTIVE</span>
                </div>
                <div class="saas-metrics-grid">
                    <div class="metric-box"><span class="metric-label">Clients</span><span class="metric-val">{{ store.totalMembers || 0 }}</span></div>
                    <div class="metric-box"><span class="metric-label">Activation</span><span class="metric-val">{{ store.account_activated ? '✅ Done' : '⏳ Pending' }}</span></div>
                    <div class="metric-box"><span class="metric-label">Staff</span><span class="metric-val">1 / {{ store.max_users || 1 }}</span></div>
                    <div class="metric-box"><span class="metric-label">Billing</span><span class="metric-val">{{ store.billing_cycle }}</span></div>
                </div>
                <div class="actions" style="margin-top:15px; display:flex; gap:10px;">
                    <button class="btn-small" style="flex:1;" @click="openEdit(store)">✏️ Config</button>
                    <a :href="'/?store=' + store.slug" target="_blank" class="btn-small" style="flex:1; text-align:center; background:var(--primary); color:white; border-color:var(--primary); text-decoration:none;">Dashboard ↗</a>
                </div>
            </div>
        </div>

        <div v-if="suspendedStores.length > 0" class="section-title">Suspended</div>
        <div class="grid-stores">
            <div v-for="store in suspendedStores" :key="store.id" class="card-box inactive">
                <div class="store-header"><h3>{{ store.name }}</h3><span class="status-badge status-suspended" @click="toggleStatus(store)">SUSPENDED</span></div>
                <button class="btn-primary" @click="toggleStatus(store)">Reactivate</button>
            </div>
        </div>

        <div v-if="editingStore" class="modal-overlay" @click.self="editingStore = null">
            <div class="modal" style="width:500px;">
                <div class="modal-header"><h2>Edit Tenant</h2><span class="close-btn" @click="editingStore = null">✕</span></div>
                <div class="form-group"><label>Name</label><input v-model="editingStore.name"></div>
                <div class="form-group"><label>Tier</label><select v-model="editingStore.tier"><option value="free_trial">Free</option><option value="basic">Basic</option><option value="lite">Lite</option><option value="pro">Pro</option></select></div>
                <div style="background:#f1f5f9; padding:15px; border-radius:10px; margin:20px 0;">
                    <button style="background:none; border:none; color:var(--primary); font-weight:800; cursor:pointer;" @click="emit('show-toast', 'Reset link sent!')">📧 Send Password Reset</button>
                </div>
                <button class="btn-primary" @click="saveChanges">SAVE CONFIG</button>
            </div>
        </div>
    </div>
    `,
    setup(props, { emit }) {
        const stores = Vue.ref([]);
        const search = Vue.ref('');
        const editingStore = Vue.ref(null);

        const fetchStores = async () => { const res = await fetch('/api/v1/master/stores'); if(res.ok) stores.value = await res.json() || []; };
        const filtered = Vue.computed(() => { if(!stores.value) return []; return stores.value.filter(s => s.name.toLowerCase().includes(search.value.toLowerCase())); });
        const activeStores = Vue.computed(() => filtered.value.filter(s => s.isActive));
        const suspendedStores = Vue.computed(() => filtered.value.filter(s => !s.isActive));

        const toggleStatus = async (s) => {
            const res = await fetch('/api/v1/master/stores/status', { method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({id: s.id, isActive: !s.isActive}) });
            if(res.ok) { fetchStores(); emit('show-toast', "Status Updated"); }
        };
        const saveChanges = async () => { const res = await fetch('/api/v1/master/stores/update', { method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(editingStore.value) }); if(res.ok) { editingStore.value = null; fetchStores(); emit('show-toast', "Saved"); } };

        Vue.onMounted(fetchStores);
        return { stores, search, activeStores, suspendedStores, editingStore, toggleStatus, openEdit: (s) => editingStore.value = {...s}, saveChanges, emit };
    }
};