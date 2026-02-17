const MasterCardSkins = {
    template: `
    <div class="skins-view" style="max-width: 1200px; margin: 0 auto; padding-bottom: 50px;">
        
        <div class="header-bar" style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 40px;">
            <div class="page-title">
                <h2 style="margin: 0; font-size: 1.8rem; font-weight: 800; color: #111827;">Global Card Skins</h2>
                <p style="margin: 5px 0 0; color: #6b7280; font-size: 0.95rem;">Design, configure and distribute themes.</p>
            </div>
            <div class="header-actions" style="display: flex; gap: 15px; align-items: center;">
                <div class="search-container" style="position: relative;">
                    <span style="position: absolute; left: 12px; top: 50%; transform: translateY(-50%); color: #9ca3af;">🔍</span>
                    <input v-model="searchSkins" placeholder="Search themes..." style="padding: 10px 10px 10px 38px; border: 1px solid #e5e7eb; border-radius: 8px; width: 250px; font-size: 0.9rem; outline: none;">
                </div>
                <button @click="openNewSkinModal" style="padding: 10px 20px; background: #2563eb; color: white; border: none; border-radius: 8px; font-weight: 600; cursor: pointer; font-size: 0.9rem; box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05); display: flex; align-items: center; gap: 5px;">
                    <span>＋</span> New Theme
                </button>
            </div>
        </div>
        
        <div style="margin-bottom: 50px;">
            <div style="font-size: 0.8rem; font-weight: 800; text-transform: uppercase; color: #111827; margin-bottom: 15px; letter-spacing: 0.5px;">SYSTEM CORE (PROTECTED)</div>
            <div style="width: 100%; height: 1px; background: #e5e7eb; margin-bottom: 25px;"></div>
            
            <div class="grid-stores" style="display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 25px;">
                <div v-for="skin in systemSkins" :key="skin.id" class="card-box" style="background: white; border: 1px solid #e5e7eb; border-radius: 12px; padding: 20px; transition: all 0.2s; box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1);">
                    <div :style="{backgroundColor: skin.colorBg}" style="height: 180px; border-radius: 10px; position: relative; overflow: hidden; margin-bottom: 15px; border: 1px solid rgba(0,0,0,0.05);">
                        <img v-if="skin.image" :src="skin.image" style="position: absolute; top:0; left:0; width:100%; height:100%; object-fit: cover;" :style="getImgTransform(skin, 3)">
                        <div style="position: relative; z-index: 10; display: flex; flex-direction: column; justify-content: space-between; height: 100%; padding: 20px; box-sizing: border-box;">
                            <div style="display: flex; justify-content: space-between; font-weight: 800; font-size: 0.7rem; text-shadow: 0 1px 2px rgba(0,0,0,0.1);">
                                <span :style="{color: skin.colorText}">VIP</span> <span :style="{color: skin.colorText}">2 🎁</span>
                            </div>
                            <div style="display: flex; gap: 5px; justify-content: center;">
                                <div v-for="n in 5" style="width: 30px; height: 30px; border-radius: 50%; border: 2px dashed rgba(255,255,255,0.5);" :style="{borderColor: skin.colorBorder}"></div>
                            </div>
                            <div style="font-weight: 800; font-size: 0.8rem; text-shadow: 0 1px 2px rgba(0,0,0,0.1);" :style="{color: skin.colorText}">USER</div>
                        </div>
                    </div>
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                        <div>
                            <h3 style="margin: 0; font-size: 1rem; color: #111827; font-weight: 700;">{{ skin.name }}</h3>
                            <p style="margin: 2px 0 0; font-size: 0.75rem; color: #6b7280;">✨ System Default</p>
                        </div>
                        <span style="font-size: 0.65rem; font-weight: 700; color: #9ca3af; border: 1px solid #e5e7eb; padding: 4px 8px; border-radius: 4px; background: #f9fafb;">LOCKED</span>
                    </div>
                </div>
            </div>
        </div>

        <div style="margin-bottom: 50px;">
            <div style="font-size: 0.8rem; font-weight: 800; text-transform: uppercase; color: #111827; margin-bottom: 10px; letter-spacing: 0.5px;">STANDARD COLLECTION</div>
            
            <div style="display: flex; align-items: center; margin-bottom: 25px;">
                <span style="font-size: 0.7rem; font-weight: 800; color: #10b981; letter-spacing: 1px; margin-right: 15px;">ACTIVE</span>
                <div style="flex: 1; height: 1px; background: #e5e7eb;"></div>
            </div>

            <div class="grid-stores" style="display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 25px; margin-bottom: 40px;">
                <div v-for="skin in sortedStandardSkins.active" :key="skin.id" class="card-box" style="background: white; border: 1px solid #e5e7eb; border-radius: 12px; padding: 20px; transition: all 0.2s; box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1);">
                    <div class="skin-preview-list" :style="getSkinStyle(skin)" style="height: 190px; border-radius: 8px; position: relative; overflow: hidden; margin-bottom: 15px; border: 1px solid rgba(0,0,0,0.05);">
                        <img v-if="skin.image" :src="skin.image" class="skin-preview-bg" :style="getImgTransform(skin, 3)" style="position: absolute; top:0; left:0; width:100%; height:100%; object-fit: cover;">
                        <div style="position: absolute; top: 10px; right: 10px; display: flex; flex-direction: column; gap: 5px; align-items: flex-end; z-index: 20;">
                            <span v-if="skin.isGlobal" style="padding: 4px 8px; border-radius: 6px; font-size: 0.6rem; font-weight: 800; text-transform: uppercase; background: rgba(255,255,255,0.95); color: #2563eb; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">Global</span>
                            <span v-if="skin.storeId" style="padding: 4px 8px; border-radius: 6px; font-size: 0.6rem; font-weight: 800; text-transform: uppercase; background: rgba(255,255,255,0.95); color: #7c3aed; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">{{ getStoreName(skin.storeId) }}</span>
                        </div>
                        <div class="card-ui" style="position: relative; z-index: 10; display: flex; flex-direction: column; justify-content: space-between; height: 100%; padding: 20px;">
                            <div class="overlay-top" style="display: flex; justify-content: space-between; font-weight: 800; font-size: 0.7rem; text-shadow: 0 1px 2px rgba(0,0,0,0.3);">
                                <span :style="{color: skin.colorText}">VIP</span> <span :style="{color: skin.colorText}">2 🎁</span>
                            </div>
                            <div class="mini-stamps-grid" style="display: flex; justify-content: space-between; gap: 5px;">
                                <div v-for="n in 5" style="width: 32px; height: 32px; border-radius: 50%; border: 2px dashed rgba(255,255,255,0.6);" :style="{borderColor: skin.colorBorder}"></div>
                            </div>
                            <div class="overlay-bottom" :style="{color: skin.colorText}" style="font-weight: 800; font-size: 0.8rem; text-shadow: 0 1px 2px rgba(0,0,0,0.3);">USER</div>
                        </div>
                    </div>
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                        <h3 style="margin: 0; font-size: 1rem; color: #111827; font-weight: 700;">{{ skin.name }}</h3>
                        <div style="display: flex; gap: 8px;">
                            <button @click="editSkin(skin)" style="width: 32px; height: 32px; border-radius: 50%; border: 1px solid #e5e7eb; background: white; cursor: pointer; display: flex; align-items: center; justify-content: center; color: #4b5563; transition: all 0.2s;">✏️</button>
                            <button @click="deleteSkin(skin.id)" style="width: 32px; height: 32px; border-radius: 50%; border: 1px solid #e5e7eb; background: white; cursor: pointer; display: flex; align-items: center; justify-content: center; color: #ef4444; transition: all 0.2s;">🗑️</button>
                        </div>
                    </div>
                </div>
            </div>

            <div style="display: flex; align-items: center; margin-bottom: 25px;">
                <span style="font-size: 0.7rem; font-weight: 800; color: #9ca3af; letter-spacing: 1px; margin-right: 15px;">DRAFTS</span>
                <div style="flex: 1; height: 1px; background: #e5e7eb;"></div>
            </div>

            <div class="grid-stores" style="display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap: 25px;">
                <div v-for="skin in sortedStandardSkins.inactive" :key="skin.id" class="card-box" style="background: #f9fafb; border: 1px dashed #d1d5db; border-radius: 12px; padding: 20px; transition: all 0.2s;">
                    <div class="skin-preview-list" :style="getSkinStyle(skin)" style="height: 190px; border-radius: 8px; position: relative; overflow: hidden; margin-bottom: 15px; border: 1px solid rgba(0,0,0,0.05); opacity: 0.85;">
                        <img v-if="skin.image" :src="skin.image" class="skin-preview-bg" :style="getImgTransform(skin, 3)" style="position: absolute; top:0; left:0; width:100%; height:100%; object-fit: cover;">
                        
                        <div style="position: absolute; top: 10px; right: 10px; z-index: 20;">
                            <span style="padding: 4px 8px; border-radius: 6px; font-size: 0.6rem; font-weight: 800; text-transform: uppercase; background: #e5e7eb; color: #6b7280; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">DRAFT</span>
                        </div>

                        <div class="card-ui" style="position: relative; z-index: 10; display: flex; flex-direction: column; justify-content: space-between; height: 100%; padding: 20px;">
                            <div class="overlay-top" style="display: flex; justify-content: space-between; font-weight: 800; font-size: 0.7rem; text-shadow: 0 1px 2px rgba(0,0,0,0.1);">
                                <span :style="{color: skin.colorText}">VIP</span> <span :style="{color: skin.colorText}">2 🎁</span>
                            </div>
                            <div class="mini-stamps-grid" style="display: flex; justify-content: space-between; gap: 5px;">
                                <div v-for="n in 5" style="width: 32px; height: 32px; border-radius: 50%; border: 2px dashed rgba(255,255,255,0.6);" :style="{borderColor: skin.colorBorder}"></div>
                            </div>
                            <div class="overlay-bottom" :style="{color: skin.colorText}" style="font-weight: 800; font-size: 0.8rem; text-shadow: 0 1px 2px rgba(0,0,0,0.1);">USER</div>
                        </div>
                    </div>
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                        <h3 style="margin: 0; font-size: 1rem; color: #6b7280; font-weight: 700;">{{ skin.name }}</h3>
                        <div style="display: flex; gap: 8px;">
                            <button @click="editSkin(skin)" style="padding: 6px 12px; border: 1px solid #d1d5db; background: white; border-radius: 6px; font-weight: 600; font-size: 0.8rem; cursor: pointer;">⚙️ Configure</button>
                            <button @click="deleteSkin(skin.id)" style="width: 32px; height: 32px; border-radius: 50%; border: 1px solid #d1d5db; background: white; cursor: pointer; display: flex; align-items: center; justify-content: center; color: #ef4444;">🗑️</button>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div style="margin-bottom: 50px;">
            <div style="font-size: 0.8rem; font-weight: 800; text-transform: uppercase; color: #111827; margin-bottom: 10px; letter-spacing: 0.5px;">SEASONAL CAMPAIGNS</div>
            
            <div style="display: flex; align-items: center; margin-bottom: 25px;">
                <span style="font-size: 0.7rem; font-weight: 800; color: #10b981; letter-spacing: 1px; margin-right: 15px;">ACTIVE</span>
                <div style="flex: 1; height: 1px; background: #e5e7eb;"></div>
            </div>

            <div class="grid-stores" style="display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap: 25px; margin-bottom: 40px;">
                <div v-for="skin in sortedSeasonalSkins.active" :key="skin.id" class="card-box" style="background: white; border: 1px solid #e5e7eb; border-radius: 12px; padding: 20px; box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1);">
                    <div class="skin-preview-list" :style="getSkinStyle(skin)" style="height: 190px; border-radius: 8px; position: relative; overflow: hidden; margin-bottom: 15px; border: 1px solid rgba(0,0,0,0.05);">
                        <img v-if="skin.image" :src="skin.image" class="skin-preview-bg" :style="getImgTransform(skin, 3)" style="position: absolute; top:0; left:0; width:100%; height:100%; object-fit: cover;">
                        
                        <div style="position: absolute; top: 10px; right: 10px; z-index: 20; display: flex; flex-direction: column; gap: 5px; align-items: flex-end;">
                            <span style="padding: 4px 8px; border-radius: 6px; font-size: 0.6rem; font-weight: 800; text-transform: uppercase; background: rgba(255,255,255,0.95); color: #f59e0b; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">Seasonal</span>
                            <span v-if="skin.isGlobal" style="padding: 4px 8px; border-radius: 6px; font-size: 0.6rem; font-weight: 800; text-transform: uppercase; background: rgba(255,255,255,0.95); color: #2563eb; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">Global</span>
                            <span v-if="skin.storeId" style="padding: 4px 8px; border-radius: 6px; font-size: 0.6rem; font-weight: 800; text-transform: uppercase; background: rgba(255,255,255,0.95); color: #7c3aed; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">{{ getStoreName(skin.storeId) }}</span>
                        </div>

                        <div class="card-ui" style="position: relative; z-index: 10; display: flex; flex-direction: column; justify-content: space-between; height: 100%; padding: 20px; box-sizing: border-box;">
                            <div class="overlay-top" style="display: flex; justify-content: space-between; font-weight: 800; font-size: 0.7rem; text-shadow: 0 1px 2px rgba(0,0,0,0.1);">
                                <span :style="{color: skin.colorText}">VIP</span> <span :style="{color: skin.colorText}">2 🎁</span>
                            </div>
                            <div class="mini-stamps-grid" style="display: flex; justify-content: space-between; gap: 5px;">
                                <div v-for="n in 5" style="width: 30px; height: 30px; border-radius: 50%; border: 2px dashed rgba(255,255,255,0.5);" :style="{borderColor: skin.colorBorder}"></div>
                            </div>
                            <div class="overlay-bottom" :style="{color: skin.colorText}" style="font-weight: 800; font-size: 0.8rem; text-shadow: 0 1px 2px rgba(0,0,0,0.1);">USER</div>
                        </div>
                    </div>
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                        <div>
                            <h3 style="margin: 0; font-size: 1rem; color: #111827; font-weight: 700;">{{ skin.name }}</h3>
                            <p style="margin: 2px 0 0; font-size: 0.75rem; color: #6b7280;">📅 {{ formatDate(skin.start) }} - {{ formatDate(skin.end) }}</p>
                        </div>
                        <div style="display: flex; gap: 8px;">
                            <button @click="editSkin(skin)" style="width: 32px; height: 32px; border-radius: 50%; border: 1px solid #e5e7eb; background: white; cursor: pointer; display: flex; align-items: center; justify-content: center; color: #4b5563;">✏️</button>
                            <button @click="deleteSkin(skin.id)" style="width: 32px; height: 32px; border-radius: 50%; border: 1px solid #e5e7eb; background: white; cursor: pointer; display: flex; align-items: center; justify-content: center; color: #ef4444;">🗑️</button>
                        </div>
                    </div>
                </div>
            </div>

            <div style="display: flex; align-items: center; margin-bottom: 25px;">
                <span style="font-size: 0.7rem; font-weight: 800; color: #9ca3af; letter-spacing: 1px; margin-right: 15px;">DRAFTS</span>
                <div style="flex: 1; height: 1px; background: #e5e7eb;"></div>
            </div>

            <div class="grid-stores" style="display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap: 25px;">
                <div v-for="skin in sortedSeasonalSkins.inactive" :key="skin.id" class="card-box" style="background: #f9fafb; border: 1px dashed #d1d5db; border-radius: 12px; padding: 20px; opacity: 0.8;">
                    <div class="skin-preview-list" :style="getSkinStyle(skin)" style="height: 190px; border-radius: 8px; position: relative; overflow: hidden; margin-bottom: 15px; border: 1px solid rgba(0,0,0,0.05); opacity: 0.85;">
                        <img v-if="skin.image" :src="skin.image" class="skin-preview-bg" :style="getImgTransform(skin, 3)" style="position: absolute; top:0; left:0; width:100%; height:100%; object-fit: cover;">
                        
                        <div style="position: absolute; top: 10px; right: 10px; z-index: 20;">
                            <span style="padding: 4px 8px; border-radius: 6px; font-size: 0.6rem; font-weight: 800; text-transform: uppercase; background: #e5e7eb; color: #6b7280; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">DRAFT</span>
                        </div>

                        <div class="card-ui" style="position: relative; z-index: 10; display: flex; flex-direction: column; justify-content: space-between; height: 100%; padding: 20px; box-sizing: border-box;">
                            <div class="overlay-top" style="display: flex; justify-content: space-between; font-weight: 800; font-size: 0.7rem; text-shadow: 0 1px 2px rgba(0,0,0,0.1);">
                                <span :style="{color: skin.colorText}">VIP</span> <span :style="{color: skin.colorText}">2 🎁</span>
                            </div>
                            <div class="mini-stamps-grid" style="display: flex; justify-content: space-between; gap: 5px;">
                                <div v-for="n in 5" style="width: 30px; height: 30px; border-radius: 50%; border: 2px dashed rgba(255,255,255,0.5);" :style="{borderColor: skin.colorBorder}"></div>
                            </div>
                            <div class="overlay-bottom" :style="{color: skin.colorText}" style="font-weight: 800; font-size: 0.8rem; text-shadow: 0 1px 2px rgba(0,0,0,0.1);">USER</div>
                        </div>
                    </div>
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                        <h3 style="margin: 0; font-size: 1rem; color: #6b7280; font-weight: 700;">{{ skin.name }}</h3>
                        <div style="display: flex; gap: 8px;">
                            <button @click="editSkin(skin)" style="padding: 6px 12px; border: 1px solid #d1d5db; background: white; border-radius: 6px; font-weight: 600; font-size: 0.8rem; cursor: pointer;">⚙️ Configure</button>
                            <button @click="deleteSkin(skin.id)" style="width: 32px; height: 32px; border-radius: 50%; border: 1px solid #d1d5db; background: white; cursor: pointer; display: flex; align-items: center; justify-content: center; color: #ef4444;">🗑️</button>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div v-if="skinBuilder" style="position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.6); display: flex; align-items: center; justify-content: center; z-index: 1000; backdrop-filter: blur(4px);" @click.self="skinBuilder = null">
            <div style="background: white; width: 850px; padding: 40px; border-radius: 20px; box-shadow: 0 25px 50px -12px rgba(0,0,0,0.25); max-height: 95vh; overflow-y: auto;">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 30px; padding-bottom: 20px; border-bottom: 1px solid #e5e7eb;">
                    <h2 style="margin: 0; font-size: 1.5rem; font-weight: 800; color: #111827;">Theme Designer</h2>
                    <span @click="skinBuilder = null" style="cursor: pointer; font-size: 1.5rem; color: #6b7280;">✕</span>
                </div>
                
                <div style="display: grid; grid-template-columns: 320px 1fr; gap: 50px;">
                    <div>
                        <div style="margin-bottom: 20px;">
                            <label style="display: block; font-size: 0.7rem; font-weight: 800; margin-bottom: 8px; color: #6b7280; text-transform: uppercase;">THEME NAME</label>
                            <input type="text" v-model="skinBuilder.name" placeholder="Ex: Summer 2026" style="width: 100%; padding: 12px; border: 1px solid #e5e7eb; border-radius: 8px; font-size: 0.95rem; outline: none; box-sizing: border-box;">
                        </div>
                        
                        <div style="margin-bottom: 20px;">
                            <label style="display: block; font-size: 0.7rem; font-weight: 800; margin-bottom: 8px; color: #6b7280; text-transform: uppercase;">AVAILABILITY</label>
                            <select v-model="availabilityMode" style="width: 100%; padding: 12px; border: 1px solid #e5e7eb; border-radius: 8px; font-size: 0.95rem; background: white; box-sizing: border-box;">
                                <option value="global">🌍 Global Network</option>
                                <option value="store">🔒 Specific Store</option>
                                <option value="draft">📝 Draft (Inactive)</option>
                            </select>
                        </div>

                        <div v-if="availabilityMode === 'store'" style="margin-bottom: 20px;">
                            <label style="display: block; font-size: 0.7rem; font-weight: 800; margin-bottom: 8px; color: #6b7280; text-transform: uppercase;">SELECT STORE</label>
                            <select v-model="skinBuilder.storeId" style="width: 100%; padding: 12px; border: 1px solid #e5e7eb; border-radius: 8px; font-size: 0.95rem; background: white; box-sizing: border-box;">
                                <option v-for="s in stores" :key="s.id" :value="s.id">{{ s.name }}</option>
                            </select>
                        </div>

                        <div style="margin-bottom: 20px;">
                            <label style="display: block; font-size: 0.7rem; font-weight: 800; margin-bottom: 8px; color: #6b7280; text-transform: uppercase;">TYPE</label>
                            <select v-model="skinBuilder.type" style="width: 100%; padding: 12px; border: 1px solid #e5e7eb; border-radius: 8px; font-size: 0.95rem; background: white; box-sizing: border-box;">
                                <option value="standard">Standard</option>
                                <option value="seasonal">Seasonal</option>
                            </select>
                        </div>

                        <div v-if="skinBuilder.type === 'seasonal'" style="background: #fff7ed; padding: 15px; border: 1px solid #ffedd5; border-radius: 8px;">
                            <div style="margin-bottom: 15px;">
                                <label style="color: #9a3412; font-size: 0.7rem; font-weight: 800; display: block; margin-bottom: 5px;">START DATE</label>
                                <input type="date" v-model="skinBuilder.start" style="width: 100%; padding: 8px; border: 1px solid #fed7aa; border-radius: 6px; box-sizing: border-box;">
                            </div>
                            <div style="margin-bottom: 0;">
                                <label style="color: #9a3412; font-size: 0.7rem; font-weight: 800; display: block; margin-bottom: 5px;">END DATE</label>
                                <input type="date" v-model="skinBuilder.end" style="width: 100%; padding: 8px; border: 1px solid #fed7aa; border-radius: 6px; box-sizing: border-box;">
                            </div>
                        </div>
                    </div>

                    <div>
                        <div style="display: block; font-size: 0.7rem; font-weight: 800; margin-bottom: 15px; color: #6b7280; text-transform: uppercase;">PREVIEW (DRAG IMAGE TO POSITION)</div>
                        
                        <div :style="{backgroundColor: skinBuilder.colorBg}" style="width: 100%; aspect-ratio: 1.58/1; border-radius: 16px; position: relative; overflow: hidden; cursor: grab; box-shadow: 0 10px 30px -5px rgba(0,0,0,0.1);" @mousedown="startDrag" @touchstart="startDrag">
                            <img v-if="skinBuilder.image" :src="skinBuilder.image" style="position: absolute; top:0; left:0; transform-origin: top left; pointer-events: none;" :style="getImgTransform(skinBuilder, 1)">
                            <div style="position: absolute; inset: 0; padding: 30px; display: flex; flex-direction: column; justify-content: space-between; pointer-events: none; z-index: 10;">
                                <div style="display: flex; justify-content: space-between; font-weight: 800; font-size: 0.85rem; letter-spacing: 1px;">
                                    <span :style="{color: skinBuilder.colorText}">BRAND VIP</span>
                                    <span :style="{color: skinBuilder.colorText}">2 🎁</span>
                                </div>
                                <div style="display: grid; grid-template-columns: repeat(5, 1fr); gap: 15px;">
                                    <div v-for="n in 10" style="aspect-ratio: 1; border-radius: 50%; border: 2px dashed; opacity: 0.7;" :style="{borderColor: skinBuilder.colorBorder}"></div>
                                </div>
                                <div>
                                    <div :style="{color: skinBuilder.colorText}" style="font-weight: 800; font-size: 1rem;">JOHN DOE</div>
                                    <div :style="{color: skinBuilder.colorText, opacity: 0.7, fontSize: '0.65rem', fontWeight: 600, marginTop: '2px'}">MEMBER Nº 0000</div>
                                </div>
                            </div>
                        </div>

                        <div style="display: grid; grid-template-columns: 1.5fr 1fr; gap: 20px; align-items: center; margin-top: 25px; margin-bottom: 25px;">
                            <label style="display: flex; align-items: center; justify-content: center; gap: 8px; width: 100%; padding: 12px; border: 1px dashed #2563eb; background: #eff6ff; color: #2563eb; border-radius: 8px; cursor: pointer; font-weight: 700; font-size: 0.8rem; box-sizing: border-box;">
                                <input type="file" accept="image/*" @change="handleImageUpload" hidden>
                                {{ skinBuilder.image ? '✅ IMAGE LOADED' : '📂 UPLOAD IMAGE' }}
                            </label>
                            <div style="display: flex; align-items: center; gap: 10px; font-size: 0.7rem; font-weight: 800; color: #6b7280; text-transform: uppercase;">
                                ZOOM: <input type="range" min="0.5" max="3" step="0.1" v-model.number="skinBuilder.scale" style="width: 100%; accent-color: #2563eb;">
                            </div>
                        </div>

                        <div style="display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 15px;">
                            <div>
                                <label style="display: block; font-size: 0.65rem; font-weight: 800; margin-bottom: 6px; color: #6b7280;">BACKGROUND</label>
                                <div :style="{backgroundColor: skinBuilder.colorBg}" style="height: 45px; width: 100%; border-radius: 8px; border: 1px solid #e5e7eb; overflow: hidden; position: relative; cursor: pointer; box-shadow: 0 1px 2px rgba(0,0,0,0.05);">
                                    <input type="color" v-model="skinBuilder.colorBg" style="position: absolute; top: -50%; left: -50%; width: 200%; height: 200%; cursor: pointer; border: none; padding: 0;">
                                </div>
                            </div>
                            <div>
                                <label style="display: block; font-size: 0.65rem; font-weight: 800; margin-bottom: 6px; color: #6b7280;">TEXT COLOR</label>
                                <div :style="{backgroundColor: skinBuilder.colorText}" style="height: 45px; width: 100%; border-radius: 8px; border: 1px solid #e5e7eb; overflow: hidden; position: relative; cursor: pointer; box-shadow: 0 1px 2px rgba(0,0,0,0.05);">
                                    <input type="color" v-model="skinBuilder.colorText" style="position: absolute; top: -50%; left: -50%; width: 200%; height: 200%; cursor: pointer; border: none; padding: 0;">
                                </div>
                            </div>
                            <div>
                                <label style="display: block; font-size: 0.65rem; font-weight: 800; margin-bottom: 6px; color: #6b7280;">STAMP BORDER</label>
                                <div :style="{backgroundColor: skinBuilder.colorBorder}" style="height: 45px; width: 100%; border-radius: 8px; border: 1px solid #e5e7eb; overflow: hidden; position: relative; cursor: pointer; box-shadow: 0 1px 2px rgba(0,0,0,0.05);">
                                    <input type="color" v-model="skinBuilder.colorBorder" style="position: absolute; top: -50%; left: -50%; width: 200%; height: 200%; cursor: pointer; border: none; padding: 0;">
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                
                <button @click="saveSkin" style="width: 100%; padding: 16px; background: #2563eb; color: white; border: none; border-radius: 10px; font-weight: 700; font-size: 1rem; margin-top: 35px; cursor: pointer; box-shadow: 0 4px 6px -1px rgba(37, 99, 235, 0.2); transition: background 0.2s;">SAVE THEME</button>
            </div>
        </div>

        <div v-if="previewSkinData" style="position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.6); display: flex; align-items: center; justify-content: center; z-index: 1000; backdrop-filter: blur(4px);" @click.self="previewSkinData = null">
            <div style="background: white; width: 450px; padding: 30px; border-radius: 16px; text-align: center;">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
                    <h2 style="margin:0; font-size:1.2rem; font-weight:800; color:#111827;">Preview</h2>
                    <span @click="previewSkinData = null" style="cursor:pointer; font-size:1.5rem; color:#6b7280;">✕</span>
                </div>
                <div :style="{backgroundColor: previewSkinData.colorBg || '#cbd5e1'}" style="width:100%; aspect-ratio:1.58/1; border-radius:12px; position:relative; overflow:hidden; margin-bottom:20px;">
                    <img v-if="previewSkinData.image" :src="previewSkinData.image" style="position:absolute; top:0; left:0; pointer-events:none;" :style="getImgTransform(previewSkinData, 1)">
                    <div style="position:absolute; inset:0; padding:20px; display:flex; flex-direction:column; justify-content:space-between;">
                        <div style="display:flex; justify-content:space-between; font-weight:800; font-size:0.75rem;"><span :style="{color: previewSkinData.colorText}">BRAND VIP</span><span :style="{color: previewSkinData.colorText}">2 🎁</span></div>
                        <div style="display:grid; grid-template-columns:repeat(5, 1fr); gap:10px;"><div v-for="n in 10" style="aspect-ratio:1; border-radius:50%; border:2px dashed; opacity:0.7;" :style="{borderColor: previewSkinData.colorBorder}"></div></div>
                        <div style="font-weight:800; font-size:0.9rem; text-align:left;"><span :style="{color: previewSkinData.colorText}">JOHN DOE</span></div>
                    </div>
                </div>
                <button @click="previewSkinData = null" style="width:100%; padding:12px; background:#2563eb; color:white; border:none; border-radius:8px; font-weight:700;">CLOSE</button>
            </div>
        </div>
    </div>
    `,
    setup(props, { emit }) {
        const { ref, computed, onMounted } = Vue;
        const globalSkins = ref([]);
        const searchSkins = ref('');
        const stores = ref([]);
        const skinBuilder = ref(null);
        const previewSkinData = ref(null);
        const availabilityMode = ref('global');
        
        // Drag Logic State
        const isDragging = ref(false);
        const startX = ref(0); const startY = ref(0);
        const initialImgX = ref(0); const initialImgY = ref(0);

        const showToast = (msg, type) => emit('show-toast', msg, type);

        const fetchSkins = async () => { 
            try { 
                const res = await fetch('/api/v1/master/skins'); 
                if(res.ok) {
                    const data = await res.json() || [];
                    globalSkins.value = data.map(s => ({
                        ...s,
                        colorBg: s.colorBg || '#cbd5e1', 
                        colorText: s.colorText || '#ffffff', 
                        colorBorder: s.colorBorder || '#ffd166',
                        scale: s.scale || 1, 
                        x: s.x || 0, 
                        y: s.y || 0
                    }));
                }
            } catch(e) { console.error(e); }
        };

        const fetchStores = async () => { 
            try { 
                const res = await fetch('/api/v1/master/stores'); 
                if(res.ok) stores.value = await res.json() || []; 
            } catch(e) { console.error(e); }
        };

        const matchesSearch = (s) => !searchSkins.value || (s.name || '').toLowerCase().includes(searchSkins.value.toLowerCase());

        const systemSkins = computed(() => globalSkins.value.filter(s => ['default', 'black'].includes(s.id) && matchesSearch(s)));
        
        const organizeSkins = (type) => {
            const list = globalSkins.value.filter(s => s.type === type && !['default', 'black'].includes(s.id) && matchesSearch(s));
            return {
                active: list.filter(s => s.isGlobal || s.storeId),
                inactive: list.filter(s => !s.isGlobal && !s.storeId)
            };
        };

        const sortedStandardSkins = computed(() => organizeSkins('standard'));
        const sortedSeasonalSkins = computed(() => organizeSkins('seasonal'));

        const openNewSkinModal = () => {
            skinBuilder.value = { 
                id: '', name: '', type: 'standard', 
                image: null, x: 0, y: 0, scale: 1, 
                colorBg: '#cbd5e1', colorText: '#ffffff', colorBorder: '#ffd166',
                start: null, end: null
            };
            availabilityMode.value = 'global';
        };
        
        const editSkin = (skin) => {
            let s = skin.start ? skin.start.split('T')[0] : '';
            let e = skin.end ? skin.end.split('T')[0] : '';
            skinBuilder.value = { ...skin, start: s, end: e };
            if (skin.isGlobal) availabilityMode.value = 'global';
            else if (skin.storeId) availabilityMode.value = 'store';
            else availabilityMode.value = 'draft';
        };

        const deleteSkin = async (id) => {
            if(!confirm("Delete this skin?")) return;
            const res = await fetch(`/api/v1/master/skins?id=${id}`, { method: 'DELETE' });
            if(res.ok) { showToast("Skin deleted", "success"); fetchSkins(); }
        };
        
        const previewSkin = (skin) => previewSkinData.value = skin;

        const handleImageUpload = (e) => {
            const file = e.target.files[0];
            if(!file) return;
            const reader = new FileReader();
            reader.onload = (ev) => { skinBuilder.value.image = ev.target.result; };
            reader.readAsDataURL(file);
        };

        const startDrag = (e) => {
            if (!skinBuilder.value.image) return;
            isDragging.value = true;
            const clientX = e.touches ? e.touches[0].clientX : e.clientX;
            const clientY = e.touches ? e.touches[0].clientY : e.clientY;
            startX.value = clientX; startY.value = clientY;
            initialImgX.value = skinBuilder.value.x || 0; 
            initialImgY.value = skinBuilder.value.y || 0;
            
            document.addEventListener('mousemove', onDrag); 
            document.addEventListener('mouseup', stopDrag);
            document.addEventListener('touchmove', onDrag); 
            document.addEventListener('touchend', stopDrag);
        };

        const onDrag = (e) => {
            if (!isDragging.value) return;
            const clientX = e.touches ? e.touches[0].clientX : e.clientX;
            const clientY = e.touches ? e.touches[0].clientY : e.clientY;
            skinBuilder.value.x = initialImgX.value + (clientX - startX.value);
            skinBuilder.value.y = initialImgY.value + (clientY - startY.value);
        };

        const stopDrag = () => {
            isDragging.value = false;
            document.removeEventListener('mousemove', onDrag);
            document.removeEventListener('mouseup', stopDrag);
            document.removeEventListener('touchmove', onDrag); 
            document.removeEventListener('touchend', stopDrag);
        };

        const saveSkin = async () => {
            const payload = { ...skinBuilder.value };
            
            if (availabilityMode.value === 'global') { payload.isGlobal = true; payload.storeId = null; } 
            else if (availabilityMode.value === 'store') { payload.isGlobal = false; } 
            else { payload.isGlobal = false; payload.storeId = null; } // Draft

            if (!payload.start) payload.start = null; else payload.start = new Date(payload.start).toISOString();
            if (!payload.end) payload.end = null; else payload.end = new Date(payload.end).toISOString();

            const res = await fetch('/api/v1/master/skins', { 
                method: 'POST', 
                headers: {'Content-Type': 'application/json'}, 
                body: JSON.stringify(payload) 
            });

            if(res.ok) { 
                showToast("Theme Saved!", "success"); 
                skinBuilder.value = null;
                fetchSkins(); 
            } else { 
                showToast("Error saving theme", "error"); 
            }
        };

        const getSkinStyle = (skin) => { 
            if (!skin) return { backgroundColor: '#cbd5e1' };
            return { backgroundColor: skin.colorBg || '#cbd5e1' }; 
        };
        const getImgTransform = (s, div) => ({ transform: `translate(${s.x/div}px, ${s.y/div}px) scale(${s.scale})` });
        const getStoreName = (id) => { const s = stores.value.find(x => x.id === id); return s ? s.name : 'Unknown'; };
        const formatDate = (d) => d ? new Date(d).toLocaleDateString('en-GB') : '';
        const checkExpiring = (dateStr) => {
            if(!dateStr) return false;
            const end = new Date(dateStr);
            const now = new Date();
            const diff = (end - now) / (1000 * 60 * 60 * 24);
            return diff < 5 && diff > 0;
        };

        Vue.onMounted(() => { fetchSkins(); fetchStores(); });

        return { 
            globalSkins, searchSkins, stores, skinBuilder, previewSkinData, availabilityMode, 
            systemSkins, sortedStandardSkins, sortedSeasonalSkins,
            openNewSkinModal, editSkin, deleteSkin, previewSkin, handleImageUpload, startDrag, saveSkin,
            getSkinStyle, getImgTransform, getStoreName, formatDate, checkExpiring
        };
    }
};