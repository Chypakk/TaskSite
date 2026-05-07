
import { GroupsService } from '../services/GroupsService.js';

export class GroupModal{

    constructor() {
        this.isInitialized = false;
        this.form = null;
        this.bootstrapModal = null;
        this.groupService = new GroupsService(); 
        this.groupsCash = null;
        this.dialogMode = {
            create: 0,
            select: 1,
            edit: 2,
        };
        this.curentMode = this.dialogMode.create;
    }


    initialize() {
        if (this.isInitialized) return;
        
        this.modalElement = document.getElementById('groupModal');
        this.form = document.getElementById('groupForm');


        if (!this.modalElement|| !this.form) {
            console.error('Tasks modal element not found!');
            return;
        }

        // Проверяем, что Bootstrap загрузился
        if (typeof bootstrap === 'undefined') {
            console.error('Bootstrap не загружен! Проверь подключение bootstrap.bundle.min.js');
            return;
        }

        // Инициализируем Bootstrap modal
        this.bootstrapModal = new bootstrap.Modal(this.modalElement, {
            backdrop: true, // Затемнение фона
            keyboard: true, // Закрытие по ESC
            focus: true
        });

        this.bindEvents();
        this.isInitialized = true;

        console.log('GroupModal initialized successfully');
    }

    bindEvents() {
        // Обработчики форм
        this.form.addEventListener('submit', (e) => this.handleSubmit(e));
        
        // Очистка при закрытии
        this.modalElement.addEventListener('hidden.bs.modal', () => {
            this.clearForm();
        });

        // Кнопка создания
        const createBtn = document.getElementById('createGroupBtn');
        if (createBtn) {
            createBtn.addEventListener('click', () => this.openCreateMode());
        }

        document.getElementById("groups").addEventListener('change', (e) => {
            const selectedGroupData = this.groupsCash.find(a => a.group_id == e.target.value);

            document.getElementById('groupName').value = selectedGroupData.group_name;
            document.getElementById('groupDescription').value = selectedGroupData.group_desc;
        });

        // Открытие редактора группы
        document.addEventListener('group:edit', (e) => {
            this.openEditMode(e.detail.groupId);
        });
    }

    // Открытие в режиме создания
    openCreateMode() {
        this.curentMode = this.dialogMode.create;
        this.clearForm();
        this.disabledElements(false);
        // Заголовок
        document.getElementById('groupModalTitle').innerHTML = 
            '<i class="fas fa-file-alt me-2"></i>Новая группа';
        this.bootstrapModal.show();
    }

    // Открытие в режиме выбора
    async openSelectMode(taskId) {
        this.curentMode = this.dialogMode.select;
        this.clearForm();
        
        // Показываем загрузку
        this.setLoading(true);
        this.disabledElements(true);
        try {

            await this.uppdateGroupCash();
            // Заполняем форму
            document.getElementById('taskId').value = taskId;

            const selectGroup = document.getElementById('groups');
            selectGroup.innerHTML = '<option value="" disabled selected>Выберите группу</option>'
            + this.groupsCash.map(g => `<option value="${g.group_id}">${g.group_name}</option>`).join('');

            // Заголовок
            document.getElementById('groupModalTitle').innerHTML = 
                `<i class="fas fa-edit me-2"></i>Выберите группу`;
            
            this.bootstrapModal.show();
        } catch (error) {
            this.showError('Не удалось загрузить данные');
        } finally {
            this.setLoading(false);
        }
    }

    // Открытие в режиме изменения
    async openEditMode(groupId) {
        this.curentMode = this.dialogMode.edit;
        this.clearForm();
        
        // Показываем загрузку
        this.setLoading(true);
        this.disabledElements(false);
        try {
            await this.uppdateGroupCash();

            document.getElementById('groupId').value = groupId;
            const selectedGroupData = this.groupsCash.find(a => a.group_id == groupId);
            document.getElementById('groupName').value = selectedGroupData.group_name;
            document.getElementById('groupDescription').value = selectedGroupData.group_desc;

            // Заголовок
            document.getElementById('groupModalTitle').innerHTML = 
                `<i class="fas fa-edit me-2"></i>Изменение группы`;


            this.bootstrapModal.show();
        } catch (error) {
            this.showError('Не удалось загрузить данные');
        } finally {
            this.setLoading(false);
        }
    }


    async handleSubmit(e) {
        e.preventDefault();

        // Валидация Bootstrap
        if (!this.form.checkValidity()) {
            this.form.classList.add('was-validated');
            return;
        }
        
        this.setLoading(true);
        this.clearAlerts();

        let groupId = 0;

        if(document.getElementById('groupId').value != ""){
            groupId = Number.parseInt(document.getElementById('groupId').value);
        }
        else{
            groupId = Number.parseInt(document.getElementById('groups').value);
        }            

        // Собираем данные
        const formData = {
            group_id:  groupId,
            group_name: document.getElementById('groupName').value.trim(),
            group_desc: document.getElementById('groupDescription').value?.trim()
        };
        
        // Валидация длины заголовка
        if (formData.group_name.length < 3) {
            this.showError('Наименование должно быть не менее 3 символов');
            this.setLoading(false);
            return;
        }
        
        try {
            // Определяем метод и URL
            let response = null;
            switch(this.curentMode){

                case this.dialogMode.create:
                    await this.groupService.createGroup(formData);
                    this.showSuccess('Группа создана!');
                break;

                case this.dialogMode.select:
                    const taskId = document.getElementById('taskId').value;

                    response = await this.groupService.putTaskInGroup(taskId,formData);
                    if(response.ok){
                        this.showSuccess('Заявка добавлена в группу!');
                    }
                    else{
                        this.showError(error.message || 'Ошибка при сохранении');
                        this.setLoading(false);
                        return;
                    }
                break;

                case this.dialogMode.edit:
                   
                    response = await this.groupService.editGroup(groupId,formData);
                    if(response.ok){
                        this.showSuccess('Группа изменена!');
                    }
                    else{
                        this.showError(error.message || 'Ошибка при сохранении');
                        this.setLoading(false);
                        return;
                    }
                break;

            }
            
            // Закрываем через 1 секунду
            setTimeout(() => {
                this.bootstrapModal.hide();
                // Событие для обновления таблицы
                document.dispatchEvent(new CustomEvent('task:saved'));
            }, 1000);
            
        } catch (error) {
            this.showError(error.message || 'Ошибка при сохранении');
        } finally {
            this.setLoading(false);
        }
    }

    // Утилиты

    async uppdateGroupCash(){
        this.groupsCash = await this.groupService.getAllGroups();
    }

    disabledElements(block) {
        document.getElementById('groupName').disabled = block;
        document.getElementById('groupDescription').disabled = block;

        const groupList = document.getElementById('groupsList');
        if(block){
            groupList.classList.remove("element_hidden");
        }
        else{
            groupList.classList.add("element_hidden");
        }     
    }

    clearForm() {
        this.form.reset();
        this.form.classList.remove('was-validated');
        document.getElementById('taskId').value = '';
        document.getElementById('groupId').value = '';
        this.clearAlerts();
    }

    clearAlerts() {
        document.getElementById('groupErrorAlert').style.display = 'none';
        document.getElementById('groupSuccessAlert').style.display = 'none';
    }

    showError(message) {
        document.getElementById('groupErrorText').textContent = message;
        document.getElementById('groupErrorAlert').style.display = 'block';
        document.getElementById('groupSuccessAlert').style.display = 'none';
        // Автоскрытие через 5 секунд
        setTimeout(() => {
            this.clearAlerts();
        }, 5000);
    }
    
    showSuccess(message) {
        document.getElementById('groupSuccessText').textContent = message;
        document.getElementById('groupSuccessAlert').style.display = 'block';
        document.getElementById('groupErrorAlert').style.display = 'none';
    }
    
    setLoading(isLoading) {
        const btn = document.getElementById('groupSaveBtn');
        if (isLoading) {
            btn.disabled = true;
            btn.innerHTML = '<i class="fas fa-spinner fa-spin me-2"></i>Сохранение...';
        } else {
            btn.disabled = false;
            btn.innerHTML = '<i class="fas fa-save me-2"></i>Сохранить';
        }
    }
}

