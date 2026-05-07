import { TasksService } from '../services/TasksService.js';
import * as FormatService from '../services/FormatService.js';

export class TasksTable{

    constructor() {
        this.isInitialized = false;
        this.tasksService = new TasksService();
        this.isFetching = false;
        
    }

    initialize() {
        if (this.isInitialized) return;
        
        this.bindEvents();
        this.isInitialized = true;

        console.log('TasksTable initialized successfully');
    }

    bindEvents() {

        document.addEventListener('task:saved', () => {
            this.fetchData(); // Перезагружаем таблицу
        });

        // Обновление таблицы после удаления
        document.addEventListener('task:deleted', () => {
            this.fetchData();
        });
        
        document.getElementById('tasksTableBody').addEventListener('click', (e)=>{
            this.selectTaskGroup(e);
        });
    }
    
    async fetchData(showLoading = true, status) {
        if (this.isFetching) return;
        this.isFetching = true;
            
        try {
            if (showLoading) {
                document.getElementById('tasksTableBody').innerHTML = `
                    <tr><td colspan="16" style="text-align: center; padding: 20px;">
                        <div class="loading"></div> Обновление данных...
                    </td></tr>
                `;
            }
            
            const data = await this.tasksService.getAllTasks(status);
            // this.renderTable(data);
            this.newRenderTable(data);

        } catch (error) {
            console.error('Ошибка загрузки данных:', error);
            document.getElementById('tasksTableBody').innerHTML = `
                <tr><td colspan="16" style="text-align: center; padding: 20px; color: #ef4444;"> 
                Молчать! "Ошибка загрузки данных" ❌. Сейчас говорит сервер: ${error}.
                </td></tr>
            `;
        } finally {
            this.isFetching = false;
        }
    }

    newRenderTable(data) {
        const tbody = document.getElementById('tasksTableBody');
        tbody.innerHTML = '';
        
        if (data.length === 0) {
            tbody.innerHTML = `<tr><td colspan="16" style="text-align: center; padding: 20px;">Нет данных</td></tr>`;
            return;
        }
        
        const formatingData = this.formatingData(data);

        formatingData.forEach(group => {
            const groupRow = document.createElement('tr');
            groupRow.classList.add('group-header');     
            groupRow.setAttribute('data-group-id', group.id);
            groupRow.innerHTML = `
                <td class='toggle-icon'>►</td>
                <td>${group.id}</td>
                <td>${group.name}</td>
                <td>${group.desc}</td>
                <td>${group.tasks.length} задач</td>
                <td>
                    <button type="button" class="btn btn-primary btn-edit-group" data-group-id="${group.id}>
                        <i class="fas fa-edit me-2"></i>Редактировать
                    </button>
                </td>
            `;
            tbody.appendChild(groupRow);

            //Добавляем обработчик на кнопку редактирования
            const editBtn = groupRow.querySelector('.btn-edit-group');
            if (editBtn) {
                editBtn.addEventListener('click', (e) => {
                    e.stopPropagation();  //Останавливаем всплытие события!

                    document.dispatchEvent(new CustomEvent('group:edit', { 
                        detail: { groupId: group.id } 
                    }));
                });
            }

            // Создаем строку-контейнер
            const tasksContainer = document.createElement('tr');     
            tasksContainer.setAttribute('data-group-id', group.id);
            tasksContainer.classList.add('group-content', 'd-none');

            const cell = document.createElement('td');
            cell.setAttribute('colspan', '6');  // По количеству колонок в шапке группы
            cell.classList.add('p-0');  // Убираем отступы (Bootstrap класс)

            const tasksTable = document.createElement("table");
            tasksTable.classList.add('table', 'table-hover', 'align-middle', 'mb-0');
            tasksTable.style.cssText = `
                width: 95%;
                margin-left: auto;
            `;
            tasksTable.innerHTML = `
                        <thead class="table-light">
                            <tr>
                                <th scope="col"> </th>
                                <th scope="col">№</th>
                                <th scope="col">Название</th>
                                <th scope="col">Автор</th>
                                <th scope="col">Статус</th>
                                <th scope="col">Взята в работу</th>
                                <th scope="col">Дата создания</th>
                                <th scope="col">Дата изменения</th>
                                <th scope="col">Завершена</th>
                            </tr>
                        </thead>`;

            cell.appendChild(tasksTable);      // Таблицу вставляем в td
            tasksContainer.appendChild(cell);  // td вставляем в tr
            tbody.appendChild(tasksContainer); // Вставляем строку с таблицей в таблицу под группу

            const tasksTableBody = document.createElement('tbody');
            tasksTable.appendChild(tasksTableBody);
            group.tasks.forEach(item => {
                const row = document.createElement('tr');
                
                if(item.status == 'completed'){
                    row.classList.add('closed-task');
                }

                row.setAttribute('data-task-id', item.id);
                // Формирование ячеек с применением классов для стилизации
                row.innerHTML = `
                    <td></td>
                    <td>${item.id}</td>
                    <td>${item.name}</td>
                    <td>${item.author}</td>
                    <td>${FormatService.getStatusText(item.status)}</td>
                    <td>${item.username == null? "-": item.username}</td>
                    <td>${FormatService.formatDate(item.created_at)}</td>
                    <td>${FormatService.formatDate(item.updated_at)}</td>
                    <td>${FormatService.formatDate(item.completed_at)}</td>
                `;
                // Клик по строке
                row.addEventListener('click', (e) => {
                    // Игнорируем клик по кнопкам действий (если они есть)
                    if (e.target.closest('.btn')) return;
                    
                    // Генерируем событие или сразу открываем
                    document.dispatchEvent(new CustomEvent('task:view', { 
                        detail: { taskId: item.id } 
                    }));
                });
                tasksTableBody.appendChild(row);
            });
        });

    }

    formatingData(data){
        const formatingData = new Map();

        for(const task of data){
            const key = task.group_id ?? 'Без номера';
            
            if(!formatingData.has(key)){
                formatingData.set(key,{
                    id: key,
                    name: key === 'Без номера' ? 'Не групированные задачи': task.group_name,
                    desc: task.group_desc || '',
                    tasks: []
                });
            }
            formatingData.get(key).tasks.push(task);
        }
        return Array.from(formatingData.values());
    }

    selectTaskGroup(event){
        const headerRow = event.target.closest('.group-header');

        if(!headerRow){
            return;
        }

        const contentRow = headerRow.nextElementSibling;
        if(!contentRow || !contentRow.classList.contains('group-content')){
            return;
        }

        const isHidden = contentRow.classList.contains('d-none');
        //const iconCell = headerRow.querySelector('.toggle-icon');
        if(isHidden){
            //добавить смену иконки iconCell
            contentRow.classList.remove('d-none');
        }
        else{
            //добавить смену иконки iconCell
            contentRow.classList.add('d-none');
        }
    }

}