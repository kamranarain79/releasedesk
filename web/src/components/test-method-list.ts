import {html, LitElement} from 'lit';
import {customElement, state} from 'lit/decorators.js';

@customElement('test-method-list')
export class TestMethodList extends LitElement {
  override createRenderRoot() {
    return this;
  }

  @state()
  extraCount = 1;

  addMethod() {
    this.extraCount++;
  }

  override render() {
    return html`
      ${Array.from({length: this.extraCount}, () => html`
        <div class="mb-4 rounded-md border border-dashed border-gray-300 p-4 space-y-3">
          <input type="text" name="Title"
                 placeholder="Title (e.g. RAUC install on device)"
                 class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 sm:text-sm">
          <textarea name="Steps" rows="3" placeholder="Steps"
                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 sm:text-sm"></textarea>
          <input type="text" name="ExpectedResult"
                 placeholder="Expected result"
                 class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 sm:text-sm">
        </div>
      `)}
      <button type="button"
              class="text-sm/6 font-semibold text-violet-800 hover:text-violet-700"
              @click="${this.addMethod}">
        <span aria-hidden="true">+</span> Add another test method
      </button>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'test-method-list': TestMethodList;
  }
}
