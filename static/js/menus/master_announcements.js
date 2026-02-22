const MasterAnnouncements = {
    template: `
    <div class="stores-view" style="max-width: 800px; margin: 0 auto; padding-bottom: 50px;">
        
        <div class="header-bar" style="margin-bottom: 30px;">
            <div class="page-title">
                <h2 style="margin: 0; font-size: 1.8rem; font-weight: 800; color: #111827;">System Announcements</h2>
                <p style="margin: 5px 0 0; color: #6b7280; font-size: 0.95rem;">Envie notificações diretamente para o painel de gestão das Lojas.</p>
            </div>
        </div>

        <div class="card-box" style="background: white; border: 1px solid #e5e7eb; border-radius: 16px; padding: 35px; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05);">
            <form @submit.prevent="sendMessage">
                
                <div class="form-group" style="margin-bottom: 20px;">
                    <label style="display: block; font-size: 0.75rem; font-weight: 800; margin-bottom: 8px; color: #374151; text-transform: uppercase;">1. Destinatário</label>
                    <select v-model="form.store_id" style="width: 100%; padding: 12px; border: 1px solid #d1d5db; border-radius: 8px; font-size: 0.95rem; outline: none; background: #f9fafb;">
                        <option value="all">📢 Todas as Lojas (Broadcast Global)</option>
                        <optgroup label="Lojas Individuais">
                            <option v-for="s in stores" :key="s.id" :value="s.id">{{ s.name }} ({{ s.admin_email }})</option>
                        </optgroup>
                    </select>
                </div>

                <div style="display: grid; grid-template-columns: 1fr 2fr; gap: 20px; margin-bottom: 20px;">
                    <div class="form-group">
                        <label style="display: block; font-size: 0.75rem; font-weight: 800; margin-bottom: 8px; color: #374151; text-transform: uppercase;">2. Tipo de Mensagem</label>
                        <select v-model="form.type" style="width: 100%; padding: 12px; border: 1px solid #d1d5db; border-radius: 8px; font-size: 0.95rem; outline: none;">
                            <option value="info">🔵 Informação Geral</option>
                            <option value="success">🟢 Novidade / Sucesso</option>
                            <option value="warning">🟠 Aviso / Pagamentos</option>
                            <option value="error">🔴 Urgente / Erro</option>
                        </select>
                    </div>

                    <div class="form-group">
                        <label style="display: block; font-size: 0.75rem; font-weight: 800; margin-bottom: 8px; color: #374151; text-transform: uppercase;">3. Título (Assunto)</label>
                        <input type="text" v-model="form.title" required placeholder="Ex: Nova funcionalidade disponível!" style="width: 100%; padding: 12px; border: 1px solid #d1d5db; border-radius: 8px; font-size: 0.95rem; outline: none; box-sizing: border-box;">
                    </div>
                </div>

                <div class="form-group" style="margin-bottom: 30px;">
                    <label style="display: block; font-size: 0.75rem; font-weight: 800; margin-bottom: 8px; color: #374151; text-transform: uppercase;">4. Mensagem Completa</label>
                    <textarea v-model="form.message" required rows="6" placeholder="Escreva a sua mensagem aqui..." style="width: 100%; padding: 15px; border: 1px solid #d1d5db; border-radius: 8px; font-size: 0.95rem; outline: none; box-sizing: border-box; resize: vertical; font-family: 'Inter', sans-serif;"></textarea>
                </div>

                <button type="submit" :disabled="isSending" style="width: 100%; padding: 16px; background: #2563eb; color: white; border: none; border-radius: 8px; font-weight: 800; font-size: 1rem; cursor: pointer; transition: 0.2s; box-shadow: 0 4px 10px rgba(37, 99, 235, 0.2); display: flex; justify-content: center; gap: 10px; align-items: center;" :style="{ opacity: isSending ? 0.7 : 1 }">
                    <span v-if="!isSending">🚀 Enviar Mensagem</span>
                    <span v-else>A enviar...</span>
                </button>
            </form>
        </div>
    </div>
    `,
    setup(props, { emit }) {
        const { ref, onMounted } = Vue;
        
        const stores = ref([]);
        const isSending = ref(false);
        const form = ref({
            store_id: 'all',
            title: '',
            message: '',
            type: 'info'
        });

        // Vai buscar as lojas para preencher o dropdown
        const fetchStores = async () => {
            try {
                const res = await fetch('/api/v1/master/stores');
                if (res.ok) stores.value = await res.json() || [];
            } catch (e) { console.error(e); }
        };

        const sendMessage = async () => {
            if (!confirm(`Tem a certeza que quer enviar esta mensagem para ${form.value.store_id === 'all' ? 'TODAS AS LOJAS' : 'esta loja específica'}?`)) return;

            isSending.value = true;
            try {
                const res = await fetch('/api/v1/master/notifications/send', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(form.value)
                });

                if (res.ok) {
                    emit('show-toast', "Mensagem enviada com sucesso!", "success");
                    form.value.title = '';
                    form.value.message = '';
                } else {
                    emit('show-toast', "Erro ao enviar a mensagem.", "error");
                }
            } catch (e) {
                emit('show-toast', "Erro de ligação.", "error");
            } finally {
                isSending.value = false;
            }
        };

        onMounted(fetchStores);

        return { stores, form, isSending, sendMessage };
    }
};